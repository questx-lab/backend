package model

import (
	"time"
)

type ReceiptMessage struct {
	ReceiptStatus uint64    `json:"receipt_status"`
	TxHash        string    `json:"tx_hash"`
	BlockHeight   int64     `json:"block_height"`
	Timestamp     time.Time `json:"timestamp"`
	TxStatus      uint64    `json:"receipt_status"`
}
