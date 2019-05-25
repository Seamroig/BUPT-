package main

import "fmt"
import "log"

func (cli *CLI) getBalance(address string) {
	if !ValidateAddress(address) {
		log.Panic("地址错误")
	}

	bc := NewBlockChain(address) //根据地址创建
	defer bc.db.Close()          //延迟关闭数据库
	balance := 0
	pubkeyhash := Base58Decode([]byte(address)) //提取公钥
	pubkeyhash = pubkeyhash[1 : len(pubkeyhash)-4]
	UTXOs := bc.FindUTXO(pubkeyhash) //查找交易金额
	for _, out := range UTXOs {
		balance += out.Value //取出金额

	}
	fmt.Printf("查询金额如下%s ：%d \n", address, balance)
}
