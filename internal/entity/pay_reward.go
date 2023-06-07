package entity

import (
	"database/sql"

	"github.com/questx-lab/backend/pkg/enum"
)

type PayRewardStatusType string

var (
	PayRewardPending    = enum.New(PayRewardStatusType("pending"))
	PayRewardInProgress = enum.New(PayRewardStatusType("inprogress"))
	PayRewardSuccess    = enum.New(PayRewardStatusType("success"))
	PayRewardFailure    = enum.New(PayRewardStatusType("failure"))
)

type PayReward struct {
	Base

	UserID string
	User   User `gorm:"foreignKey:UserID"`

	ClaimedQuestID sql.NullString
	ClaimedQuest   ClaimedQuest `gorm:"foreignKey:ClaimedQuestID"`

	// Note contains the reason of this transaction in case of not come from a
	// claimed quest.
	Note    string
	Status  PayRewardStatusType
	Address string
	Token   string
	Amount  float64

	TxHash string
}
