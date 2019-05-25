package main

import (
	"fmt"
	"strconv"
)

func main() {
	fmt.Println("Welcome to my own cryptocurrency system")
	fmt.Println("欢迎使用我的加密货币——Vcoin")
	fmt.Println()
	bc := NewBlockchain() //创建区块链
	bc.AddBlock("挖矿时间太久了")
	bc.AddBlock("等不了了")

	for _, block := range bc.blocks {
		fmt.Printf("上一块哈希: %x   ", block.PrevBlockHash)
		fmt.Printf("\n数据: %s    ", block.Data)
		fmt.Printf("\n当前哈希%x", block.Hash)
		pow := NewProofOfWork(block) //校验工作量
		fmt.Printf("\npow %s\n", strconv.FormatBool(pow.Validate()))

		fmt.Println()
		fmt.Println()
	}
}
