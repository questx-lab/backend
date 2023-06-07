package entity

type BlockChainTransaction struct {
	Chain       string
	TxHash      string
	BlockHeight int64
	TxBytes     []byte
}
