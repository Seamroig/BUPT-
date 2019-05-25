package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64 //最大的64位整数

)

const targetBits = 16 //对比的位数，如果越大那么难度也会越大

type ProofOfWork struct {
	block  *Block   //区块
	target *big.Int //存储计算哈希对比的特定整数
}

//创建一个工作量证明的挖矿对象

func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)                  //初始化目标整数
	target.Lsh(target, uint(256-targetBits)) //数据转化
	pow := &ProofOfWork{block, target}       //创建对象
	return pow
}

//准备数据进行挖矿

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,       //上一块哈希
			pow.block.Data,                //当前数据
			IntToHex(pow.block.Timestamp), //时间十六进制
			IntToHex(int64(targetBits)),   //位数，十六进制
			IntToHex(int64(nonce)),        //保存工作量的nonce
		}, []byte{},
	)

	return data
}

//进行挖矿
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	fmt.Printf("当前挖矿计算的数据%s", pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce) //准备好的数据
		hash = sha256.Sum256(data)     //计算哈希
		fmt.Printf("\r%x", hash)       //打印哈希
		hashInt.SetBytes(hash[:])      //获取要对比的数据
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++ //挖矿的校验
		}

	}
	fmt.Println("\n\n")
	return nonce, hash[:] //nonce是答案

}

//校验挖矿是否成功
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.prepareData(pow.block.Nonce) //准备好的数据
	hash := sha256.Sum256(data)              //计算出哈希
	hashInt.SetBytes(hash[:])                //获取要对比的数据
	isValid := (hashInt.Cmp(pow.target) == -1)

	return isValid
}
