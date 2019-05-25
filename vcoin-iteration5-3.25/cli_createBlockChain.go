package main


import "log"
import "fmt"


func (cli *CLI)createBlockChain(address string){
	if !ValidateAddress(address){
		log.Panic("地址错误")
	
	}
	bc:=CreateBlockChain(address)		//创建一个区块链
	bc.db.Close()
	fmt.Println("创建成功")	
}

