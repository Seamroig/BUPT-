package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

//命令行接口
type CLI struct {
	blockchain *BlockChain
}

//用法
func (cli *CLI) printUsage() {
	fmt.Println("使用方法如下")
	fmt.Println("addblock  向区块链增加块")
	fmt.Println("showchain  显示区块链")

}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage() //显示用法
		os.Exit(1)
	}
}

func (cli *CLI) addBlock(data string) {
	cli.blockchain.AddBlock(data) //增加区块
	fmt.Println("区块增加成功")
}

func (cli *CLI) showBlockChain() {
	bci := cli.blockchain.Iterator() //创建循环迭代器
	for {
		block := bci.next() //取得下一个区块

		fmt.Printf("上一块哈希: %x   ", block.PrevBlockHash)
		fmt.Println()
		fmt.Printf("数据: %s    ", block.Data)
		fmt.Println()
		fmt.Printf("当前哈希: %x", block.Hash)
		fmt.Println()
		pow := NewProofOfWork(block)
		fmt.Printf("pow %s", strconv.FormatBool(pow.Validate()))
		fmt.Println("\n")

		if len(block.PrevBlockHash) == 0 { //循环到创世区块
			break
		}
	}
}

func (cli *CLI) Run() {
	cli.validateArgs() //校验

	//处理命令行参数
	addblockcmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	showchaincmd := flag.NewFlagSet("showchain", flag.ExitOnError)

	addBlockData := addblockcmd.String("data", "", "block data")
	switch os.Args[1] {
	case "addblock":
		err := addblockcmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic(err) //处理错误
		}
	case "showchain":
		err := showchaincmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic(err) //处理错误
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}
	if addblockcmd.Parsed() {
		if *addBlockData == "" {
			addblockcmd.Usage()
			os.Exit(1)
		} else {
			cli.addBlock(*addBlockData)
		}
	}

	if showchaincmd.Parsed() {
		cli.showBlockChain() //显示区块链
	}
}
