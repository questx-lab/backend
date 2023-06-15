package entity

import "github.com/questx-lab/backend/pkg/enum"

type TxStatusType string

var (
	TxStatusTypePending    = enum.New(TxStatusType("pending"))
	TxStatusTypeInProgress = enum.New(TxStatusType("inprogress"))
	TxStatusTypeSuccess    = enum.New(TxStatusType("success"))
	TxStatusTypeFailure    = enum.New(TxStatusType("failure"))
)

type BlockChainTransaction struct {
	PayRewardID string
	Status      TxStatusType

	Token  string
	Amount float64

	Chain       string
	TxHash      string
	BlockHeight int64
	TxBytes     []byte
}
