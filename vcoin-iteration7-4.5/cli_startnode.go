package main

import "fmt"
import "log"

func (cli *CLI) startNode(nodeID, minerAddress string) {
	fmt.Printf("开启一个节点%s", nodeID)
	if len(minerAddress) > 0 {
		if ValidateAddress(minerAddress) {
			fmt.Printf("正在挖矿，地址为%s", minerAddress)
		} else {
			log.Panic("错误的挖矿地址")
		}
	}
	StartServer(nodeID, minerAddress) //开启服务器

}
