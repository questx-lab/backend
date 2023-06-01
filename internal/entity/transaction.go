package entity

import (
	"database/sql"

	"github.com/questx-lab/backend/pkg/enum"
)

type TransactionStatusType string

var (
	TransactionPending    = enum.New(TransactionStatusType("pending"))
	TransactionInProgress = enum.New(TransactionStatusType("inprogress"))
	TransactionSuccess    = enum.New(TransactionStatusType("success"))
	TransactionFailure    = enum.New(TransactionStatusType("failure"))
)

type Transaction struct {
	Base

	UserID string
	User   User `gorm:"foreignKey:UserID"`

	ClaimedQuestID sql.NullString
	ClaimedQuest   ClaimedQuest `gorm:"foreignKey:ClaimedQuestID"`

	// Note contains the reason of this transaction in case of not come from a
	// claimed quest.
	Note    string
	Status  TransactionStatusType
	Address string
	Token   string
	Amount  float64

	TxHash string
}
