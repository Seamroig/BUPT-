package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

const protocol = "tcp"   //安全的协议
const nodeVersion = 1    //版本
const commandLength = 12 //命令行长度

var nodeAddress string                     //节点地址
var miningAddress string                   //挖矿地址
var knowNodes = []string{"localhost:3000"} //已知的节点
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction) //内存池

type addr struct {
	Addrlist []string //节点
}

type block struct {
	AddrFrom string //来源地址
	Block    []byte
}

type getblocks struct {
	AddrFrom string
}

type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type tx struct {
	AddFrom     string
	Transaction []byte
}

type verzion struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

//字节到命令
func bytesToCommand(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b) //增加命令的编号
		}
	}
	return fmt.Sprintf("%s", command)
}

//命令到字节
func commandToBytes(command string) []byte {
	var bytes [commandLength]byte
	for i, c := range command {
		bytes[i] = byte(c) //字节转化为索引
	}
	return bytes[:]
}

//提取命令
func extractCommand(request []byte) []byte {
	return request[:commandLength]
}

//请求块
func requestBlocks() {
	for _, node := range knowNodes {
		sendGetBlocks(node) //给所有已经知道的节点发送请求
	}
}

//发送块
func sendBlock(addr string, bc *Block) {
	data := block{nodeAddress, bc.Serialize()}             //构建模块
	payload := gobEncode(data)                             //数据处理
	request := append(commandToBytes("block"), payload...) //定制请求
	sendData(addr, request)                                //发送数据

}

//发送地址
func sendaddr(address string) {
	nodes := addr{knowNodes} //已知的所有节点
	nodes.Addrlist = append(nodes.Addrlist, nodeAddress)
	payload := gobEncode(nodes)                           //增加解码的节点
	request := append(commandToBytes("addr"), payload...) //创建请求
	sendData(address, request)                            //发送数据

}

//发送数据
func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr) //建立tcp网络链接对象
	if err != nil {
		fmt.Printf("%s 地址不可到达\n", addr)
		var updateNodes []string
		for _, node := range knowNodes {
			if node != addr {
				updateNodes = append(updateNodes, node) //刷新节点
			}
		}
		knowNodes = updateNodes //刷新列表
		return
	}
	defer conn.Close()
	_, err = io.Copy(conn, bytes.NewReader(data)) //拷贝数据,发送
	if err != nil {
		log.Panic(err)
	}

}

//发送请求
func sendInv(address, kind string, items [][]byte) {
	inventory := inv{nodeAddress, kind, items}           //库存数据
	payload := gobEncode(inventory)                      //历史数据
	request := append(commandToBytes("inv"), payload...) //网络请求
	sendData(address, request)                           //发送数据

}

//发送请求块
func sendGetBlocks(address string) {
	payload := gobEncode(getblocks{nodeAddress}) //解码地址
	request := append(commandToBytes("getblocks"), payload...)
	sendData(address, request) //发送数据与请求

}

//发送请求数据
func sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getdata{address, kind, id}) //解码地址
	request := append(commandToBytes("getdata"), payload...)
	sendData(address, request) //发送数据与请求
}

//发送交易
func sendTx(addr string, tnx *Transaction) {
	data := tx{nodeAddress, tnx.Serialize()} //处理数据
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)
	sendData(addr, request) //发送数据与请求
}

//发送版本信息
func sendVersion(addr string, bc *BlockChain) {
	bestHeight := bc.GetBestHeight() //最后一个区块的height
	payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request) //发送数据与请求

}

func handleBlock(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:]) //取出数据
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blockData := payload.Block
	block := DeserializeBlock(blockData)
	fmt.Printf("收到一个新的区块\n")
	bc.AddBlock(block)
	fmt.Printf("增加一个区块%x\n", block.Hash)
	if len(blocksInTransit) > 0 {
		blockhash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockhash) //发送请求
		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()

	}

}

//读取网络地址
func handleaddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:]) //取出数据
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	knowNodes = append(knowNodes, payload.Addrlist...)
	fmt.Printf("已经有了%d个节点\n", len(knowNodes))
	requestBlocks() //请求区块数据

}

//请求的版本
func handleInv(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:]) //取出数据
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("收到库存 %d %s \n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items                   //历史抓取的区块
		blockhash := payload.Items[0]                     //区块哈希
		sendGetData(payload.AddrFrom, "block", blockhash) //发送请求的数据

		newInTransit := [][]byte{} //字节二维数组
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockhash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit //同步区块

	}

	if payload.Type == "tx" {
		txID := payload.Items[0]
		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

//抓取多个区块
func handleGetBlocks(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:]) //取出数据
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)

}

//抓取数据
func handleGetData(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:]) //取出数据
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}
		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]
		sendTx(payload.AddrFrom, &tx)
	}
}

//抓取交易
func handleTx(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:]) //取出数据
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction           //获取交易数据
	tx := DeserializeTransaction(txData)    //解码交易数据
	mempool[hex.EncodeToString(tx.ID)] = tx //处理交易

	//fmt.Println(nodeAddress, knowNodes[0]) //显示数据                          //first
	if nodeAddress == knowNodes[0] {
		for _, node := range knowNodes {
			if node != nodeAddress && node != payload.AddFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {

		if len(mempool) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*Transaction //交易列表
			for id := range mempool {
				tx := mempool[id]
				if bc.VertifyTransaction(&tx) { //校验交易是否伪造
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				fmt.Println("没有任何交易，等待交易\n")
				return
			}
			cbTx := NewCoinBaseTX(miningAddress, "") //创建一个地址，为这个地址挖矿
			txs = append(txs, cbTx)                  //叠加
			newBlock := bc.MineBlock(txs)            //挖矿
			UTXOSet := UTXOSet{bc}
			UTXOSet.Reindex()
			fmt.Printf("新的区块已经挖掘到\n")
			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID) //交易编号
				delete(mempool, txID)             //删除内存池
			}
			for _, node := range knowNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash}) //挖矿成功广播
				}
			}

			if len(mempool) > 0 {
				goto MineTransactions
			}

		}

	}

}

//处理版本
func handleVersion(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload verzion

	buff.Write(request[commandLength:]) //取出数据
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	mybestHeight := bc.GetBestHeight()        //抓取最好的宽度
	foreignerBestHeight := payload.BestHeight //抓取最好的宽度

	//版本同步
	if mybestHeight < foreignerBestHeight {
		sendGetBlocks(payload.AddrFrom)
	} else if mybestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}

	if !nodeIsKnow(payload.AddrFrom) {
		knowNodes = append(knowNodes, payload.AddrFrom) //判断节点是否已知
	}

}

//处理网络链接
func handleConnection(conn net.Conn, bc *BlockChain) {
	request, err := ioutil.ReadAll(conn) //处理所有网络连接
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])
	fmt.Printf("收到命令%s\n", command)

	switch command {
	case "addr":
		handleaddr(request)
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request, bc)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		handleVersion(request, bc)

	default:
		fmt.Printf("未知数据\n")
	}

	conn.Close()

}

//开启服务器
func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress //挖矿地址
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	bc := NewBlockChain(nodeID)

	if nodeAddress != knowNodes[0] {
		sendVersion(knowNodes[0], bc)
	}

	for {
		conn, err := ln.Accept() //接受请求
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc) //异步处理
	}
}

//
func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

//判断节点是否已知
func nodeIsKnow(addr string) bool {
	for _, node := range knowNodes {
		if node == addr {
			return true
		}
	}
	return false
}
