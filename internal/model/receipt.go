package model

import (
	"encoding/json"
	"time"
)

type ReceiptMessage struct {
	ReceiptStatus uint64    `json:"receipt_status"`
	TxHash        string    `json:"tx_hash"`
	BlockHeight   int64     `json:"block_height"`
	Timestamp     time.Time `json:"timestamp"`
}

func (m *ReceiptMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *ReceiptMessage) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}
