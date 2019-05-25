package main

import "fmt"
import "log"

func (cli *CLI) send(from, to string, amount int) {
	if !ValidateAddress(from) {
		log.Panic("地址错误")

	}

	if !ValidateAddress(to) {
		log.Panic("地址错误")

	}

	bc := NewBlockChain(from)
	defer bc.db.Close()
	tx := NewUTXOTransaction(from, to, amount, bc) //转账
	bc.MineBlock([]*Transaction{tx})               //挖矿确认交易,记账成功
	fmt.Println("交易成功")
}
