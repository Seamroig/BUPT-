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

func (cli *CLI) createBlockChain(address string) {
	bc := createBlockChain(address) //创建区块链
	bc.db.Close()
	fmt.Println("创建成功", address)
}

func (cli *CLI) getBalance(address string) {
	bc := NewBlockChain(address) //根据地址创建
	defer bc.db.Close()          //延迟关闭数据库
	balance := 0
	UTXOs := bc.FindUTXO(address) //查找交易金额
	for _, out := range UTXOs {
		balance += out.Value //取出金额

	}
	fmt.Printf("查询金额如下%s ：%d \n", address, balance)

}

//用法
func (cli *CLI) printUsage() {
	fmt.Println("使用方法如下")
	fmt.Println("getbalance -address 你输入的地址 根据地址查询金额")
	fmt.Println("createblockchain -address 你输入的地址 根据地址创建区块链")
	fmt.Println("send -from From -to TO -amount Amount 转账")
	fmt.Println("showchain  显示区块链")

}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage() //显示用法
		os.Exit(1)
	}
}

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockChain(from)
	defer bc.db.Close()
	tx := NewUTXOTransaction(from, to, amount, bc) //转账
	bc.MineBlock([]*Transaction{tx})               //挖矿确认交易,记账成功
	fmt.Println("交易成功")
}

func (cli *CLI) showBlockChain() {
	bc := NewBlockChain("")
	defer bc.db.Close()
	bci := bc.Iterator()
	for {
		block := bci.next()
		fmt.Printf("上一块哈希%x\n", block.PrevBlockHash)
		fmt.Printf("当前哈希%x\n", block.Hash)
		pow := NewProofOfWork(block) //工作量证明
		fmt.Printf("pow %s \n", strconv.FormatBool(pow.Validate()))

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) Run() {
	cli.validateArgs() //校验

	//处理命令行参数
	getbalancecmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createblockchaincmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendcmd := flag.NewFlagSet("send", flag.ExitOnError)
	showchaincmd := flag.NewFlagSet("showchain", flag.ExitOnError)

	getbalanceaddress := getbalancecmd.String("address", "", "查询地址")
	createblockaddress := createblockchaincmd.String("address", "", "查询地址")
	sendfrom := sendcmd.String("from", "", "谁给的")
	sendto := sendcmd.String("to", "", "给谁的")
	sendamount := sendcmd.Int("amount", 0, "金额")

	switch os.Args[1] {
	case "getbalance":
		err := getbalancecmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err) //处理错误
		}
	case "createblockchain":
		err := createblockchaincmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic(err) //处理错误
		}

	case "send":
		err := sendcmd.Parse(os.Args[2:]) //解析参数
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

	if getbalancecmd.Parsed() {
		if *getbalanceaddress == "" {
			getbalancecmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getbalanceaddress) //查询
	}

	if createblockchaincmd.Parsed() {
		if *createblockaddress == "" {
			createblockchaincmd.Usage()
			os.Exit(1)
		}
		cli.createBlockChain(*createblockaddress) //创建区块链条
	}

	if sendcmd.Parsed() {
		if *sendfrom == "" || *sendto == "" || *sendamount <= 0 {
			sendcmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendfrom, *sendto, *sendamount)
	}

	if showchaincmd.Parsed() {
		cli.showBlockChain() //显示区块链
	}
}
