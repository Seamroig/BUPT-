package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

//整数转化为16进制
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)                        //开辟内存来存字节
	err := binary.Write(buff, binary.BigEndian, num) //num转化字节写入
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes() //返回字节集合
}
