package main

import "crypto/ecdsa"
import "crypto/elliptic"
import "crypto/rand"
import "log"
import "bytes"
import "crypto/sha256"
import "golang.org/x/crypto/ripemd160"

const version = byte(0x00)      //钱包版本
const walletfile = "wallet.dat" //钱包文件
const addressChecksumlen = 4    //监测地址长度

type Wallet struct {
	PrivateKey ecdsa.PrivateKey //钱包的权限
	PublicKey  []byte           //收款地址

}

//创建钱包
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}

//创建公钥私钥
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()                              //创建加密算法
	private, err := ecdsa.GenerateKey(curve, rand.Reader) //生成私有key
	if err != nil {
		log.Panic(err)
	}

	//生成公有
	publickey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, publickey
}

//公钥的校验
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	SecondSHA := sha256.Sum256(firstSHA[:]) //两次加密
	return SecondSHA[:addressChecksumlen]
}

//公钥哈希处理
func HashPubkey(pubkey []byte) []byte {
	publicsha256 := sha256.Sum256(pubkey)     //处理公钥
	R160Hash := ripemd160.New()               //创建一个哈希算法对象
	_, err := R160Hash.Write(publicsha256[:]) //写入处理
	if err != nil {
		log.Panic(err)
	}
	publicR160Hash := R160Hash.Sum(nil)

	return publicR160Hash
}

//抓取钱包的地址
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubkey(w.PublicKey) //取得哈希值
	versionPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionPayload)               //监测版本与公钥
	fullpayload := append(versionPayload, checksum...) //叠加校验信息
	address := Base58Encode(fullpayload)
	return address //返回钱包地址
}

func ValidateAddress(address string) bool {
	publicHash := Base58Decode([]byte(address)) //解码
	actualchecksum := publicHash[len(publicHash)-addressChecksumlen:]
	version := publicHash[0]
	publicHash = publicHash[1 : len(publicHash)-addressChecksumlen]
	targetCheckSum := checksum(append([]byte{version}, publicHash...))

	return bytes.Compare(actualchecksum, targetCheckSum) == 0
}
