package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const walletFile = "wallet_%s.dat" //钱包文件

type Wallets struct {
	Wallets map[string]*Wallet //一个钱包对应一个字符串

}

//创建一个钱包，或者抓取已经存在的钱包
func NewWallets(nodeID string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)
	err := wallets.LoadFromFile(nodeID)
	return &wallets, err
}

//创建一个钱包
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet() //创建钱包
	address := fmt.Sprintf("%s", wallet.GetAddress())
	ws.Wallets[address] = wallet //保存钱包
	return address
}

//抓取所有钱包的地址
func (ws *Wallets) GetAddresses() []string {
	var addresses []string //所有钱包地址
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses //返回所有钱包地址
}

//抓取一个钱包的地址
func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

//从文件中读取钱包
func (ws *Wallets) LoadFromFile(nodeID string) error {
	mywalletfile := fmt.Sprintf(walletFile, nodeID) //生成文件地址
	if _, err := os.Stat(mywalletfile); os.IsNotExist(err) {
		return err
	}
	fileContent, err := ioutil.ReadFile(mywalletfile) //读取文件
	if err != nil {
		log.Panic(err)
	}

	//读取文件二进制并且解析
	var wallets Wallets                                     //钱包
	gob.Register(elliptic.P256())                           //注册加密解密
	decoder := gob.NewDecoder(bytes.NewReader(fileContent)) //解码
	err = decoder.Decode(&wallets)                          //解码
	if err != nil {
		log.Panic(err)
	}
	ws.Wallets = wallets.Wallets

	return nil
}

//钱包保存到文件
func (ws *Wallets) SaveToFile(nodeID string) {
	var content bytes.Buffer
	mywalletfile := fmt.Sprintf(walletFile, nodeID) //生成文件地址
	gob.Register(elliptic.P256())                   //注册加密算法
	encoder := gob.NewEncoder(&content)             //编码
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}
	err = ioutil.WriteFile(mywalletfile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
