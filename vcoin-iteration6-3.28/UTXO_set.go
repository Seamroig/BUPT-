package main

import "log"
import "encoding/hex"
import "github.com/boltdb/bolt"

const utxoBucket = "chainstate" //存储状态

//二次封装区块链
type UTXOSet struct {
	blockchain *BlockChain
}

//输出查找并且返回未曾使用的输出
func (utxo UTXOSet) FindSpendableOutpus(publickeyhash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int) //处理输出
	accumulated := 0                         //累计的金额
	db := utxo.blockchain.db                 //调用数据库

	//查询数据
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket)) //查询数据
		cur := bucket.Cursor()                  //当前的游标
		for key, value := cur.First(); key != nil; key, value = cur.Next() {
			txID := hex.EncodeToString(key)
			outs := DeserialzieOutputs(value)
			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(publickeyhash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	return accumulated, unspentOutputs
}

//查找utxo，按照公钥来查询
func (utxo UTXOSet) FindUTXO(publickeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := utxo.blockchain.db //取出数据库进行查询
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		cur := bucket.Cursor()
		for key, value := cur.First(); key != nil; key, value = cur.Next() {
			outs := DeserialzieOutputs(value) //反序列化数据库数据
			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(publickeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}

		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	return UTXOs
}

//统计交易
func (utxo UTXOSet) CountTransactions() int {
	db := utxo.blockchain.db //引用数据库
	counter := 0
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		cur := bucket.Cursor()
		for k, _ := cur.First(); k != nil; k, _ = cur.Next() {
			counter++ //叠加
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	return counter
}

//重建索引
func (utxo UTXOSet) Reindex() {
	db := utxo.blockchain.db
	bucketname := []byte(utxoBucket)
	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketname)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic()
		}
		_, err = tx.CreateBucket(bucketname)
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	UTXO := utxo.blockchain.FindUTXO()
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketname)
		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}
			err = bucket.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
}

//刷新数据
func (utxo UTXOSet) Update(block *Block) {
	db := utxo.blockchain.db
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket)) //取出数据库对象数据
		for _, tx := range block.Transactions {
			if tx.IsCoinBase() == false {
				for _, vin := range tx.Vin {
					updateOuts := TXoutputs{}            //创建集合
					outBytes := bucket.Get(vin.Txid)     //取出数据
					outs := DeserialzieOutputs(outBytes) //解码二进制数据
					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updateOuts.Outputs = append(updateOuts.Outputs, out) //叠加序列
						}
					}
					if len(updateOuts.Outputs) == 0 {
						err := bucket.Delete(vin.Txid)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := bucket.Put(vin.Txid, updateOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}
				}

			}
			newOutputs := TXoutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out) //处理叠加
			}
			err := bucket.Put(tx.ID, newOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}

		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

}
