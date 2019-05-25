package main

import "fmt"

func main() {
	fmt.Println("hello")
	bc := NewBlockchain() //创建区块链
	bc.AddBlock("我 pay 她 10")
	bc.AddBlock("我 pay 你 20")

	for _, block := range bc.blocks {
		fmt.Printf("上一块哈希%x   ", block.PrevBlockHash)
		fmt.Printf("数据: %s    ", block.Data)
		fmt.Printf("当前哈希%x", block.Hash)

		fmt.Println()
	}
}
