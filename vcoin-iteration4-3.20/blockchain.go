package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db" //数据库文件名当前目录下面
const blockBucket = "blocks"   //一个名称
const genesisCoinbaseData = "你们好"

type BlockChain struct {
	tip []byte   //二进制数据
	db  *bolt.DB //数据库

}

type BlockChainIterator struct {
	currentHash []byte   //	当前哈希
	db          *bolt.DB //数据库
}

//挖矿
func (blockchain *BlockChain) MineBlock(transactions []*Transaction) {
	var lastHash []byte //最后的哈希
	err := blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket)) //查看数据
		lastHash = bucket.Get([]byte("1"))       //取出最后区块
		return nil
	})
	if err != nil {
		log.Panic(err) //处理错误
	}
	newBlock := NewBlock(transactions, lastHash) //创建一个新的区块
	err = blockchain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		err := bucket.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err) //处理错误
		}
		err = bucket.Put([]byte("1"), newBlock.Hash) //压入保存最后一个哈希
		if err != nil {
			log.Panic(err) //处理错误
		}
		blockchain.tip = newBlock.Hash //保存上一块的哈希
		return nil
	})

}

//获取未使用输出的交易列表
func (blockchain *BlockChain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := blockchain.FindUnspentTransactions(address)
	for _, tx := range unspentTransactions { //循环所有交易
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) { //判断是否锁定
				UTXOs = append(UTXOs, out) //加入数据
			}
		}
	}

	return UTXOs
}

//查找没有花费的交易,挖矿类
func (blockchain *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction        //交易事务
	spentTXOS := make(map[string][]int) //开辟内存
	bci := blockchain.Iterator()        //迭代器
	for {
		block := bci.next()                     //循环下一个
		for _, tx := range block.Transactions { //循环每个交易
			txID := hex.EncodeToString(tx.ID) //获取交易编号
		Outputs:
			for outindex, out := range tx.Vout {
				if spentTXOS[txID] != nil {
					for _, spentOut := range spentTXOS[txID] {
						if spentOut == outindex {
							continue Outputs //循环到不等为止
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)

				}
			}
			if tx.IsCoinBase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutPutWith(address) { //判断是否可以锁定
						inTxID := hex.EncodeToString(in.Txid) //编码为字符串
						spentTXOS[inTxID] = append(spentTXOS[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 { //最后一块跳出
			break
		}
	}

	return unspentTXs
}

//查找进行转账的交易
func (blockchain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)                  //输出
	unspentTxs := blockchain.FindUnspentTransactions(address) //根据地址查所有交易
	accmulated := 0                                           //累计
Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		for outindex, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accmulated < amount {
				accmulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outindex) //序列的叠加
				if accmulated >= amount {
					break Work
				}
			}
		}
	}

	return accmulated, unspentOutputs
}

/*
//增加一个区块
func (block *BlockChain) AddBlock(data string) {
	var lastHash []byte //上一块哈希
	err := block.db.View(func(tx *bolt.Tx) error {
		block := tx.Bucket([]byte(blockBucket)) //取得数据
		lastHash = block.Get([]byte("1"))       //取得第一块

		return nil
	})

	if err != nil {
		log.Panic(err) //处理数据库打开错误
	}

	newBlock := NewBlock(data, lastHash) //创建一个新的区块
	err = block.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))               //取出
		err := bucket.Put(newBlock.Hash, newBlock.Serialize()) //压入数据
		if err != nil {
			log.Panic(err) //处理压入错误
		}

		err = bucket.Put([]byte("1"), newBlock.Hash) //压入数据
		if err != nil {
			log.Panic(err) //处理压入错误
		}
		block.tip = newBlock.Hash

		return nil
	})

}


*/

//迭代器
func (block *BlockChain) Iterator() *BlockChainIterator {
	bcit := &BlockChainIterator{block.tip, block.db}

	return bcit //创建区块链迭代器
}

//取得下一个区块
func (it *BlockChainIterator) next() *Block {
	var block *Block
	err := it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		encodeBlock := bucket.Get(it.currentHash) //抓取二进制数据
		block = DeserializeBlock(encodeBlock)     //解码
		return nil
	})

	if err != nil {
		log.Panic(err) //处理压入错误
	}

	it.currentHash = block.PrevBlockHash //哈希赋值
	return block
}

//判断数据库是否存在
func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

//新建一个区块链
func NewBlockChain(address string) *BlockChain {
	if dbExists() == false {
		fmt.Println("数据库不存在，先创建一个")
		os.Exit(1)
	}

	var tip []byte                          //存储区块链的二进制数据
	db, err := bolt.Open(dbFile, 0600, nil) //打开数据库
	if err != nil {
		log.Panic(err) //处理数据库打开错误
	}

	//处理数据更新
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket)) //按照名称打开数据库的表格
		tip = bucket.Get([]byte("1"))

		return nil
	})
	if err != nil {
		log.Panic(err) //处理数据库更新错误
	}

	bc := BlockChain{tip, db}
	return &bc
}

//创建一个区块链创建一个数据库
func createBlockChain(address string) *BlockChain {
	if dbExists() {
		fmt.Println("数据库已经存在")
		os.Exit(1)
	}

	var tip []byte                          //存储区块链的二进制数据
	db, err := bolt.Open(dbFile, 0600, nil) //打开数据库
	if err != nil {
		log.Panic(err) //处理数据库打开错误
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinBaseTX(address, genesisCoinbaseData) //创建创世区块的事务交易
		genesis := NewGenesisBlock(cbtx)                    //创建创世区块
		bucket, err := tx.CreateBucket([]byte(blockBucket))
		if err != nil {
			log.Panic(err)
		}
		err = bucket.Put(genesis.Hash, genesis.Serialize()) //存储
		if err != nil {
			log.Panic(err)
		}
		err = bucket.Put([]byte("1"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash
		return nil
	})

	bc := BlockChain{tip, db}
	return &bc

}
