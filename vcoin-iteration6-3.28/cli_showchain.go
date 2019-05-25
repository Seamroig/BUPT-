package main

import "fmt"
import "strconv"

func (cli *CLI) showBlockChain() {
	bc := NewBlockChain()
	defer bc.db.Close()

	bci := bc.Iterator()
	for {
		block := bci.next()
		fmt.Printf("上一块哈希%x\n", block.PrevBlockHash)
		fmt.Printf("当前哈希%x\n", block.Hash)
		pow := NewProofOfWork(block) //工作量证明
		fmt.Printf("pow %s \n", strconv.FormatBool(pow.Validate()))

		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}

		fmt.Println("\n\n")
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
