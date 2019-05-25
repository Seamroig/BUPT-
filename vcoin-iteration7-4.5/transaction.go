package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

const subsidy = 1000 //挖矿的block reward

/*
type TXInput struct { //输入
	Txid      []byte //Txid存储了交易的id
	Vout      int    //Vout保存该交易中的output索引
	ScriptSig string //ScriptSig仅仅保存了一个任意的用户定义的钱包
}

//检查地址是否启动事物
func (input *TXInput) CanUnlockOutPutWith(unlockingData string) bool {
	return input.ScriptSig == unlockingData
}

*/
//将其放到transaction_input中进行重写

/*
type TXOutput struct { //输出
	Value        int    //output保存了币面值
	ScriptPubkey string //用脚本语言意味着比特币可以作为智能合约平台
}

//是否可以解锁输出
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubkey == unlockingData
}

*/
//将其放到transaction_output中重写

type Transaction struct { //交易，编号，输入和输出
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

//序列化,对象到二进制
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded) //编码器
	err := enc.Encode(tx)           //编码
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes() //返回二进制
}

//反序列化,二进制到对象
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction
	decoder := gob.NewDecoder(bytes.NewReader(data)) //解码器
	err := decoder.Decode(&transaction)              //解码
	if err != nil {
		log.Panic(err)
	}
	return transaction
}

//对于交易事务进行哈希
func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize()) //取得二进制进行哈希计算
	return hash[:]
}

//签名
func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinBase() {
		return //挖矿无需签名
	}
	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("以前的交易不正确")
		}
	}
	txCopy := tx.TrimmedCopy() //拷贝没有私钥的副本
	for inID, vin := range txCopy.Vin {
		//设置签名与公钥
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		//datatoSign := fmt.Sprintf("%x\n", txCopy) //要签名的数据

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inID].Signature = signature

	}

}

//用于签名的交易事务，裁剪的副本
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})

	}
	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})

	}
	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy
}

//把对象作为字符串展示
func (tx *Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Transaction %x\n", tx.ID))
	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("input %d", i))
		lines = append(lines, fmt.Sprintf("TXID %x", input.Txid))
		lines = append(lines, fmt.Sprintf("OUT %d", input.Vout))
		lines = append(lines, fmt.Sprintf("Signature %d", input.Signature))
		lines = append(lines, fmt.Sprintf("Pubkey %d", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("out %d", i))
		lines = append(lines, fmt.Sprintf("value %d", output.Value))
		lines = append(lines, fmt.Sprintf("out %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

//签名认证
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinBase() {
		return true //如果是挖矿无需认证
	}
	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("之前的交易错误")
		}
	}
	txCopy := tx.TrimmedCopy() //拷贝
	curve := elliptic.P256()   //加密
	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash //设置公钥
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil //公钥

		r := big.Int{}
		s := big.Int{}
		siglen := len(vin.Signature) //统计签名的长度
		r.SetBytes(vin.Signature[:(siglen / 2)])
		s.SetBytes(vin.Signature[(siglen / 2):])

		x := big.Int{}
		y := big.Int{}
		keylen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keylen / 2)])
		y.SetBytes(vin.PubKey[(keylen / 2):])
		//datatoVerify := fmt.Sprintf("%x\n", txCopy) //校验

		rawPubkey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubkey, txCopy.ID, &r, &s) == false {
			return false
		}

		//txCopy.Vin[inID].PubKey = nil
	}

	return true
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

		data = fmt.Sprintf("奖励给 %s", to)
	}
	txin := TXInput{[]byte{}, -1, nil, []byte(data)} //输入奖励
	txout := NewTXOUTput(subsidy, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash() //哈希计算
	return &tx
}

//钱包处理交易
func NewUTXOTransaction(wallet *Wallet, to string, amount int, UTXOSet *UTXOSet) *Transaction {
	var inputs []TXInput   //输入
	var outputs []TXOutput //输出

	pubkeyhash := HashPubkey(wallet.PublicKey) //获取公钥
	acc, validOutputs := UTXOSet.FindSpendableOutpus(pubkeyhash, amount)
	if acc < amount {
		log.Panic("你的钱不够")
	}
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}
		for _, out := range outs {
			//输入
			input := TXInput{txID, out, nil, wallet.PublicKey}
			inputs = append(inputs, input)

		}
	}
	from := fmt.Sprintf("%s", wallet.GetAddress())

	outputs = append(outputs, *NewTXOUTput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOUTput(acc-amount, from))
	}
	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	UTXOSet.blockchain.SignTransaction(&tx, wallet.PrivateKey)
	return &tx

}
