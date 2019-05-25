package main

func main() {
	block := NewBlockChain() //创建区块链表
	defer block.db.Close()   //延迟关闭数据
	cli := CLI{block}        //创建命令行
	cli.Run()                //开启
}
