package main

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"
)

//test
//wangjieping19970407
//6D8CF1EF201EE7D749DABEFC3AA4B6AFE20D04B10086CB45B53DFCA84CD5634D
//wangjieping199704
//9E2C9285FC650C1621B07AE68F9BF4E76FF4348E3A3EC60A93E85E5AE07E8919

//input size is changed but output size is same
//挖矿是使得后几位全部为000000...

//算力即是证明

func mainx() {
	start := time.Now() //当前时间
	//循环挖矿（计算）
	for i := 0; i < 100000000; i++ {
		data := sha256.Sum256([]byte(strconv.Itoa(i)))
		fmt.Printf("%10d,%x\n", i, data)
		fmt.Printf("%s\n", string(data[len(data)-1:]))
		if string(data[len(data)-1:]) == "0" {
			usedtime := time.Since(start)
			fmt.Printf("挖矿成功 %d Ms", usedtime)
			break
		}
	}
}

//-1则是0
//-2则是00
//-3就必然是000，也就是说末尾匹配的越多挖矿计算越困难
