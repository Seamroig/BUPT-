package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

var dbFile = "blockchain.db" //数据库文件名当前目录下面
const blockBucket = "blocks" //一个名称
const genesisCoinbaseData = "你们好"

type BlockChain struct {
	tip []byte   //二进制数据
	db  *bolt.DB //数据库

}

//挖矿
func (blockchain *BlockChain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte //最后的哈希
	var lastHeight int  //最后的长度

	for _, tx := range transactions {
		if blockchain.VertifyTransaction(tx) != true {
			log.Panic("交易不正确，有错误")
		}
	}

	err := blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket)) //查看数据
		lastHash = bucket.Get([]byte("1"))       //取出最后区块

		blockData := bucket.Get(lastHash) //取出最后的区块数据
		block := DeserializeBlock(blockData)

		lastHeight = block.Height //抓取宽度

		return nil
	})
	if err != nil {
		log.Panic(err) //处理错误
	}
	newBlock := NewBlock(transactions, lastHash, lastHeight+1) //创建一个新的区块
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

	return newBlock
}

//获取未使用输出的交易列表
func (blockchain *BlockChain) FindUTXO() map[string]TXoutputs {
	UTXO := make(map[string]TXoutputs)  //新建序列
	spentTXOs := make(map[string][]int) //花掉的交易
	bci := blockchain.Iterator()        //迭代器
	for {
		block := bci.next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spendoutidx := range spentTXOs[txID] {
						if spendoutidx == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinBase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}

		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return UTXO
}

//查找没有花费的交易,挖矿类
func (blockchain *BlockChain) FindUnspentTransactions(pubkeyhash []byte) []Transaction {
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

				if out.IsLockedWithKey(pubkeyhash) {
					unspentTXs = append(unspentTXs, *tx)

				}
			}
			if tx.IsCoinBase() == false {
				for _, in := range tx.Vin {
					if in.UsesKey(pubkeyhash) { //判断是否可以锁定
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
func (blockchain *BlockChain) FindSpendableOutputs(pubkeyhash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)                     //输出
	unspentTxs := blockchain.FindUnspentTransactions(pubkeyhash) //根据地址查所有交易
	accmulated := 0                                              //累计
Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		for outindex, out := range tx.Vout {
			if out.IsLockedWithKey(pubkeyhash) && accmulated < amount {
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

//判断数据库是否存在
func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

//新建一个区块链
func NewBlockChain(nodeID string) *BlockChain {
	dbFile = fmt.Sprintf("blockchain_%s.db", nodeID)
	if dbExists(dbFile) == false {
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
func CreateBlockChain(address string, nodeID string) *BlockChain {
	dbFile := fmt.Sprintf("blockchain_%s.db", nodeID)
	if dbExists(dbFile) {
		fmt.Println("数据库已经存在")
		os.Exit(1)
	}

	var tip []byte //存储区块链的二进制数据

	cbtx := NewCoinBaseTX(address, genesisCoinbaseData) //创建创世区块的事务交易
	genesis := NewGenesisBlock(cbtx)                    //创建创世区块

	db, err := bolt.Open(dbFile, 0600, nil) //打开数据库
	if err != nil {
		log.Panic(err) //处理数据库打开错误
	}

	err = db.Update(func(tx *bolt.Tx) error {

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
	if err != nil {
		log.Panic(err)
	}

	bc := BlockChain{tip, db}
	return &bc

}

//交易签名
func (blockchain *BlockChain) SignTransaction(tx *Transaction, privatekey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		preTx, err := blockchain.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(preTx.ID)] = preTx
	}
	tx.Sign(privatekey, prevTXs)
}

func (blockchain *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	bci := blockchain.Iterator()
	for {
		block := bci.next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return Transaction{}, nil
}

func (blockchain *BlockChain) VertifyTransaction(tx *Transaction) bool {
	prevTxs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		prevTx, err := blockchain.FindTransaction(vin.Txid) //查找交易
		if err != nil {
			log.Panic(err)
		}

		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	return tx.Verify(prevTxs)
}

//抓取最后一个区块,同步
func (blockchain *BlockChain) GetBestHeight() int {
	var lastBlock Block //最后一个区块
	err := blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket)) //取出数据库对象
		lastHash := bucket.Get([]byte("1"))      //取得最后的哈希
		blockdata := bucket.Get(lastHash)        //取得上一个哈希
		lastBlock = *DeserializeBlock(blockdata)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}

//增加块
func (blockchain *BlockChain) AddBlock(block *Block) {
	err := blockchain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		blockInDb := bucket.Get(block.Hash) //判断区块是否存在
		if blockInDb != nil {
			return nil
		}
		blockData := block.Serialize() //序列化
		err := bucket.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}
		lastHash := bucket.Get([]byte("1"))   //取出数据
		lastBlockdata := bucket.Get(lastHash) //取出最后一个区块的数据
		lastBlock := DeserializeBlock(lastBlockdata)

		if block.Height > lastBlock.Height { //判断区块链的宽度
			err = bucket.Put([]byte("1"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			blockchain.tip = block.Hash

		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

//区块链中查找区块
func (blockchain *BlockChain) GetBlock(blockhash []byte) (Block, error) {
	var bc Block //要查找的区块
	err := blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket)) //取出数据库的数据对象
		blockdata := bucket.Get(blockhash)       //取出数据
		if blockdata == nil {
			return errors.New("没有找到区块")
		}

		bc = *DeserializeBlock(blockdata)
		return nil
	})
	if err != nil {
		return bc, err
	}

	return bc, nil
}

func (blockchain *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := blockchain.Iterator()
	for {
		block := bci.next()
		blocks = append(blocks, block.Hash) //查找过程
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return blocks
}
