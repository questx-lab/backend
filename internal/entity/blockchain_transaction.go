package entity

import (
	"github.com/questx-lab/backend/pkg/enum"
)

type BlockchainTransactionStatusType string

var (
	BlockchainTransactionStatusTypeInProgress = enum.New(BlockchainTransactionStatusType("inprogress"))
	BlockchainTransactionStatusTypeSuccess    = enum.New(BlockchainTransactionStatusType("success"))
	BlockchainTransactionStatusTypeFailure    = enum.New(BlockchainTransactionStatusType("failure"))
)

type BlockchainTransaction struct {
	Base

	Chain      string     `gorm:"index:idx_blockchain_transaction_chain_txhash,unique"`
	Blockchain Blockchain `gorm:"foreignKey:Chain;references:Name"`
	TxHash     string     `gorm:"index:idx_blockchain_transaction_chain_txhash,unique"`

	Status BlockchainTransactionStatusType
}
