package main

import (
	"bytes"
	"crypto/ecdsa"
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

//挖矿
func (blockchain *BlockChain) MineBlock(transactions []*Transaction) {
	var lastHash []byte //最后的哈希
	for _, tx := range transactions {
		if blockchain.VertifyTransaction(tx) != true {
			log.Panic("交易不正确，有错误")
		}
	}

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
func (blockchain *BlockChain) FindUTXO(pubkeyhash []byte) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := blockchain.FindUnspentTransactions(pubkeyhash)
	for _, tx := range unspentTransactions { //循环所有交易
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubkeyhash) { //判断是否锁定
				UTXOs = append(UTXOs, out) //加入数据
			}
		}
	}

	return UTXOs
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
func CreateBlockChain(address string) *BlockChain {
	if dbExists() {
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

//钱包处理交易
func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TXInput         //输入
	var outputs []TXOutput       //输出
	wallets, err := NewWallets() //创建钱包
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)          //获取钱包
	pubkeyhash := HashPubkey(wallet.PublicKey) //获取公钥
	acc, validOutputs := bc.FindSpendableOutputs(pubkeyhash, amount)
	if acc < amount {
		log.Panic("你的钱不够")
	}
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}
		for _, out := range outs {
			//输入
			input := TXInput{txID, out, nil, wallet.PublicKey}
			inputs = append(inputs, input)

		}
	}

	outputs = append(outputs, *NewTXOUTput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOUTput(acc-amount, from))
	}
	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	bc.SignTransaction(&tx, wallet.PrivateKey)
	return &tx

}
