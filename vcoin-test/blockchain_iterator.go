package main

import "log"
import "github.com/boltdb/bolt"

type BlockChainIterator struct {
	currentHash []byte   //	当前哈希
	db          *bolt.DB //数据库
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
