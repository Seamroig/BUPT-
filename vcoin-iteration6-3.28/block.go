package main

import (
	/*"bytes"
	"crypto/sha256"
	"strconv"*/
	//这三个包现在并未使用了
	"bytes"
	//"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

//定义一个区块

type Block struct {
	Timestamp int64 //时间线，1970.1.1到现在有多少时间
	//Data          []byte //存我们的交易数据
	Transactions  []*Transaction //交易的集合
	PrevBlockHash []byte         //上一个块的哈希
	Hash          []byte         //当前区块的哈希
	Nonce         int            //工作量证明
}

/*
func (block *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(block.Timestamp, 10))                           //处理时间将其转化为字符
	headers := bytes.Join([][]byte{block.PrevBlockHash, block.Data, timestamp}, []byte{}) //headers用于拼成想要转化为hash的数据
	hash := sha256.Sum256(headers)                                                        //用headers计算出hash地址
	block.Hash = hash[:]
}

*/

//创建一个创世区块

func NewGenesisBlock(coinbase *Transaction) *Block {

	return NewBlock([]*Transaction{coinbase}, []byte{})

}

//对交易实现哈希计算
func (block *Block) HashTransactions() []byte {
	var transactions [][]byte
	for _, tx := range block.Transactions {
		transactions = append(transactions, tx.Serialize()) //叠加
	}
	mTree := NewMerkleTree(transactions)
	return mTree.RootNode.data
}

//创建一个区块
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0} //block是一个指针，取得一个对象初始化以后的地址
	pow := NewProofOfWork(block)                                                 //挖矿附加这个区块
	nonce, hash := pow.Run()                                                     //开始挖矿
	block.Hash = hash[:]
	block.Nonce = nonce

	//block.SetHash() //设置当前哈希
	//现在开始算出哈希而不是设置

	return block

}

//对象转化为二进制字节集，可以写入文件
func (block *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result) //开辟内存存放字节集合，之后编码
	err := encoder.Encode(block)       //编码对象创建
	if err != nil {
		log.Panic(err) //处理错误
	}
	return result.Bytes() //返回字节
}

//读取文件，读到二进制字节集，二进制字节集转化为对象
func DeserializeBlock(data []byte) *Block {
	var block Block                                  //对象存储用于字节转化的对象
	decoder := gob.NewDecoder(bytes.NewReader(data)) //解码
	err := decoder.Decode(&block)                    //尝试解码
	if err != nil {
		log.Panic(err) //处理错误
	}
	return &block
}
