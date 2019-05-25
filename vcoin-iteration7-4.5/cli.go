package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

//命令行接口
type CLI struct {
	blockchain *BlockChain
}

//用法
func (cli *CLI) printUsage() {
	fmt.Println("\n")
	fmt.Println("欢迎使用我的区块链加密货币 Vcoin")
	fmt.Println("Welcome to my Vcoin system")
	fmt.Println("\n")
	fmt.Println("使用方法如下")
	fmt.Println("listaddresses                      显示所有账户")
	fmt.Println("createwallet                       创建钱包")
	fmt.Println("getbalance -address                你输入的地址 根据地址查询金额")
	fmt.Println("createblockchain -address          你输入的地址 根据地址创建区块链")
	fmt.Println("send -from From(转出的地址) -to TO（转入的地址） -amount Amount（输入金额）  转账 -mine")
	fmt.Println("showchain                          显示区块链")
	fmt.Println("reindexutxo                        重建索引")
	fmt.Println("startnode -miner ADDR              开启一个节点")

}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage() //显示用法
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs() //校验

	//处理命令行参数
	listaddressescmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	createwalletcmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	getbalancecmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createblockchaincmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendcmd := flag.NewFlagSet("send", flag.ExitOnError)
	showchaincmd := flag.NewFlagSet("showchain", flag.ExitOnError)
	reindexutxocmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	startnodecmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	getbalanceaddress := getbalancecmd.String("address", "", "查询地址")
	createblockaddress := createblockchaincmd.String("address", "", "查询地址")
	sendfrom := sendcmd.String("from", "", "谁给的")
	sendto := sendcmd.String("to", "", "给谁的")
	sendamount := sendcmd.Int("amount", 0, "金额")
	sendmine := sendcmd.Bool("mine", false, "是否立刻挖矿")
	startnodeminer := startnodecmd.String("miner", "", "开启是否挖矿")

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

	case "listaddresses":
		err := listaddressescmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic(err) //处理错误
		}

	case "createwallet":
		err := createwalletcmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic(err) //处理错误
		}

	case "reindexutxo":
		err := reindexutxocmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic(err) //处理错误
		}
	case "startnode":
		err := startnodecmd.Parse(os.Args[2:]) //解析参数
		if err != nil {
			log.Panic(err) //处理错误
		}

	default:
		cli.printUsage()
		os.Exit(1)
	}

	var nodeID string = "3002"

	if getbalancecmd.Parsed() {
		if *getbalanceaddress == "" {
			getbalancecmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getbalanceaddress, nodeID) //查询
	}

	if createblockchaincmd.Parsed() {
		if *createblockaddress == "" {
			createblockchaincmd.Usage()
			os.Exit(1)
		}
		cli.createBlockChain(*createblockaddress, nodeID) //创建区块链条
	}

	if sendcmd.Parsed() {
		if *sendfrom == "" || *sendto == "" || *sendamount <= 0 {
			sendcmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendfrom, *sendto, *sendamount, nodeID, *sendmine)
	}

	if showchaincmd.Parsed() {
		cli.showBlockChain(nodeID) //显示区块链
	}

	if listaddressescmd.Parsed() {
		cli.listAddresses(nodeID) //列出账户
	}

	if createwalletcmd.Parsed() {
		cli.createWallet(nodeID) //创建钱包
	}

	if reindexutxocmd.Parsed() {
		cli.reindexUTXOP(nodeID) //重建索引
	}

	if startnodecmd.Parsed() {
		cli.startNode(nodeID, *startnodeminer)
	}
}
