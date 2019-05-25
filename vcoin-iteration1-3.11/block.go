package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

//定义一个区块

type Block struct {
	Timestamp     int64  //时间线，1970.1.1到现在有多少时间
	Data          []byte //存我们的交易数据
	PrevBlockHash []byte //上一个块的哈希
	Hash          []byte //当前区块的哈希
}

func (block *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(block.Timestamp, 10))                           //处理时间将其转化为字符
	headers := bytes.Join([][]byte{block.PrevBlockHash, block.Data, timestamp}, []byte{}) //headers用于拼成想要转化为hash的数据
	hash := sha256.Sum256(headers)                                                        //用headers计算出hash地址
	block.Hash = hash[:]
}

//创建一个创世区块

func NewGenesisBlock() *Block {

	return NewBlock("nimenhao", []byte{})

}

//创建一个区块

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}} //block是一个指针，取得一个对象初始化以后的地址
	block.SetHash()                                                           //设置当前哈希
	return block

}
