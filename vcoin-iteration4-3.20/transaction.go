package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const subsidy = 10 //挖矿的block reward

type TXInput struct { //输入
	Txid      []byte //Txid存储了交易的id
	Vout      int    //Vout保存该交易中的output索引
	ScriptSig string //ScriptSig仅仅保存了一个任意的用户定义的钱包
}

//检查地址是否启动事物
func (input *TXInput) CanUnlockOutPutWith(unlockingData string) bool {
	return input.ScriptSig == unlockingData
}

type TXOutput struct { //输出
	Value        int    //output保存了币面值
	ScriptPubkey string //用脚本语言意味着比特币可以作为智能合约平台
}

//是否可以解锁输出
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubkey == unlockingData
}

type Transaction struct { //交易，编号，输入和输出
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

//检查交易是否是coinbase交易
func (tx *Transaction) IsCoinBase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

//设置交易id,从二进制数据中
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer //开辟内存
	var hash [32]byte
	enc := gob.NewEncoder(&encoded) //解码对象
	err := enc.Encode(tx)           //解码
	if err != nil {
		log.Panic(err)
	}

	hash = sha256.Sum256(encoded.Bytes()) //计算哈希
	tx.ID = hash[:]                       //设置id

}

//挖矿交易
func NewCoinBaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("挖矿奖励给 %s", to)
	}
	txin := TXInput{[]byte{}, -1, data} //输入奖励
	txout := TXOutput{subsidy, to}      //输出奖励
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	return &tx
}

//普通交易,转账
func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("交易金额不足")
	}
	for txid, outs := range validOutputs { //循环无效输出
		txID, err := hex.DecodeString(txid) //解码
		if err != nil {
			log.Panic(err) //处理错误
		}
		for _, out := range outs {
			input := TXInput{txID, out, from} //输入的交易
			inputs = append(inputs, input)    //输出的交易
		}
	}

	//交易叠加
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		//记录以后的金额
		outputs = append(outputs, TXOutput{acc - amount, from})
	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID() //设置id
	return &tx //返回交易
}
