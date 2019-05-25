package main

import "log"
import "fmt"



//提取钱包所有地址
func (cli *CLI) listAddresses() {
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}

	addresses := wallets.GetAddresses()
	for _, addr := range addresses {
		fmt.Println(addr)
	}
}
