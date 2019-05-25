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

	bc := NewBlockChain()
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, &UTXOSet) //转账
	cbTx := NewCoinBaseTX(from, "")                      //挖矿
	txs := []*Transaction{cbTx, tx}                      //交易

	newBlock := bc.MineBlock(txs)
	UTXOSet.Update(newBlock)

	fmt.Println("交易成功")
}
