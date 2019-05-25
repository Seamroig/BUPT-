package main

import "fmt"
import "log"

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !ValidateAddress(from) {
		log.Panic("地址错误")

	}

	if !ValidateAddress(to) {
		log.Panic("地址错误")

	}

	bc := NewBlockChain(nodeID) //钱包集合
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from) //查找钱包

	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet) //转账

	if mineNow {
		cbTx := NewCoinBaseTX(from, "") //挖矿
		txs := []*Transaction{cbTx, tx} //交易
		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		sendTx(knowNodes[0], tx) //发送交易等待确认
	}

	fmt.Println("交易成功")
}
