package entity

import (
	"database/sql"

	"github.com/questx-lab/backend/pkg/enum"
)

type ClaimedQuestStatus string

var (
	Pending      = enum.New(ClaimedQuestStatus("pending"))
	Accepted     = enum.New(ClaimedQuestStatus("accepted"))
	Rejected     = enum.New(ClaimedQuestStatus("rejected"))
	AutoRejected = enum.New(ClaimedQuestStatus("auto_rejected"))
	AutoAccepted = enum.New(ClaimedQuestStatus("auto_accepted"))
)

type ClaimedQuest struct {
	Base

	QuestID string
	Quest   Quest `gorm:"foreignKey:QuestID"`

	UserID string
	User   User `gorm:"foreignKey:UserID"`

	SubmissionData string
	Status         ClaimedQuestStatus
	ReviewerID     string
	ReviewedAt     sql.NullTime
	Comment        string

	// Only for claiming quests with coin reward.
	Chain         string
	Blockchain    Blockchain `gorm:"foreignKey:Chain;references:Name"`
	WalletAddress string
}
