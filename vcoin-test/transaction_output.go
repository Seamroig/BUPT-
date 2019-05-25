package main

import "bytes"
import "encoding/gob"
import "log"

type TXOutput struct { //输出
	Value      int //output保存了币面值
	PubKeyHash []byte
}

//输出锁住的标志
func (out *TXOutput) Lock(address []byte) {
	pubkeyhash := Base58Decode(address)            //编码
	pubkeyhash = pubkeyhash[1 : len(pubkeyhash)-4] //截取有效哈希
	out.PubKeyHash = pubkeyhash                    //锁住无法被修改
}

//监测是否被key锁住
func (out *TXOutput) IsLockedWithKey(pubkeyHAsh []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubkeyHAsh) == 0
}

//创造一个输出
func NewTXOUTput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil} //输出
	txo.Lock([]byte(address))    //锁住
	return txo
}

type TXoutputs struct {
	Outputs []TXOutput
}

//对象到二进制
func (outs *TXoutputs) Serialize() []byte {
	var buff bytes.Buffer        //开辟内存
	enc := gob.NewEncoder(&buff) //创建编码器
	err := enc.Encode(outs)
	if err != nil {
		log.Panic()
	}
	return buff.Bytes()
}

//二进制数据到对象
func DeserialzieOutputs(data []byte) TXoutputs {
	var outputs TXoutputs
	dec := gob.NewDecoder(bytes.NewReader(data)) //解码
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic()
	}
	return outputs

}
