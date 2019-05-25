package main

type BlockChain struct {
	blocks []*Block //一个数组，每个元素都是指针，存储块的地址

}

//创建一个区块链
func NewBlockchain() *BlockChain {
	return &BlockChain{blocks: []*Block{NewGenesisBlock()}}
}

//增加一个区块
func (blocks *BlockChain) AddBlock(data string) {
	prevBlock := blocks.blocks[len(blocks.blocks)-1] //取出最后一个区块
	newBlock := NewBlock(data, prevBlock.Hash)       //创建一个区块
	blocks.blocks = append(blocks.blocks, newBlock)  //区块链插入新的区块
}
