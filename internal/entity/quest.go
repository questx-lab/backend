package entity

import (
	"database/sql"

	"github.com/questx-lab/backend/pkg/enum"
)

type QuestType string

var (
	// Basic quests
	QuestURL       = enum.New(QuestType("url"))
	QuestImage     = enum.New(QuestType("image"))
	QuestVisitLink = enum.New(QuestType("visit_link"))
	QuestText      = enum.New(QuestType("text"))
	QuestQuiz      = enum.New(QuestType("quiz"))
	QuestEmpty     = enum.New(QuestType("empty"))
	QuestInvite    = enum.New(QuestType("invite"))

	// Twitter quests
	QuestTwitterFollow    = enum.New(QuestType("twitter_follow"))
	QuestTwitterReaction  = enum.New(QuestType("twitter_reaction"))
	QuestTwitterTweet     = enum.New(QuestType("twitter_tweet"))
	QuestTwitterJoinSpace = enum.New(QuestType("twitter_join_space"))

	// Discord quests
	QuestJoinDiscord   = enum.New(QuestType("join_discord"))
	QuestInviteDiscord = enum.New(QuestType("invite_discord"))

	// Telegram quests
	QuestJoinTelegram = enum.New(QuestType("join_telegram"))
)

type RecurrenceType string

var (
	Once    = enum.New(RecurrenceType("once"))
	Daily   = enum.New(RecurrenceType("daily"))
	Weekly  = enum.New(RecurrenceType("weekly"))
	Monthly = enum.New(RecurrenceType("monthly"))
)

type QuestStatusType string

var (
	QuestDraft    = enum.New(QuestStatusType("draft"))
	QuestActive   = enum.New(QuestStatusType("active"))
	QuestArchived = enum.New(QuestStatusType("archived"))
)

type ConditionOpType string

var (
	Or  = enum.New(ConditionOpType("or"))
	And = enum.New(ConditionOpType("and"))
)

type RewardType string

var (
	DiscordRoleReward = enum.New(RewardType("discord_role"))
	CointReward       = enum.New(RewardType("coin"))
)

type ConditionType string

var (
	QuestCondition   = enum.New(ConditionType("quest"))
	DateCondition    = enum.New(ConditionType("date"))
	DiscordCondition = enum.New(ConditionType("discord"))
)

type Reward struct {
	Type RewardType `json:"type"`
	Data Map        `json:"data"`
}

type Condition struct {
	Type ConditionType `json:"type"`
	Data Map           `json:"data"`
}

type Quest struct {
	Base

	CommunityID sql.NullString
	Community   Community `gorm:"foreignKey:CommunityID"`

	IsTemplate     bool
	Type           QuestType
	Status         QuestStatusType
	Index          int
	Title          string
	Description    []byte `gorm:"type:longtext"`
	CategoryID     sql.NullString
	Category       Category `gorm:"foreignKey:CategoryID"`
	Recurrence     RecurrenceType
	ValidationData Map
	Points         uint64
	Rewards        Array[Reward]
	ConditionOp    ConditionOpType
	Conditions     Array[Condition]
	IsHighlight    bool
}
