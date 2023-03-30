package entity

import (
	"time"

	"github.com/questx-lab/backend/pkg/enum"
)

type ClaimedQuestStatus string

var (
	Pending      = enum.New(ClaimedQuestStatus("pending"))
	Accepted     = enum.New(ClaimedQuestStatus("accepted"))
	Rejected     = enum.New(ClaimedQuestStatus("rejected"))
	AutoAccepted = enum.New(ClaimedQuestStatus("auto_accepted"))
)

type ClaimedQuest struct {
	Base

	QuestID string
	Quest   Quest `gorm:"foreignKey:QuestID"`

	UserID string
	User   User `gorm:"foreignKey:UserID"`

	Input      string
	Status     ClaimedQuestStatus
	ReviewerID string
	ReviewerAt time.Time
	Comment    string
}
