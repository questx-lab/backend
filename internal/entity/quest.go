package entity

import (
	"github.com/questx-lab/backend/pkg/enum"
)

type QuestType string

var (
	QuestVisitLink = enum.New(QuestType("visit_link"))
	QuestText      = enum.New(QuestType("text"))
	QuestQuiz      = enum.New(QuestType("quiz"))

	// Twitter quests
	QuestTwitterFollow    = enum.New(QuestType("twitter_follow"))
	QuestTwitterReaction  = enum.New(QuestType("twitter_reaction"))
	QuestTwitterTweet     = enum.New(QuestType("twitter_tweet"))
	QuestTwitterJoinSpace = enum.New(QuestType("twitter_join_space"))
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

type AwardType string

var (
	PointAward  = enum.New(AwardType("points"))
	DiscordRole = enum.New(AwardType("discord_role"))
)

type ConditionType string

var (
	QuestCondition = enum.New(ConditionType("quest"))
	DateCondition  = enum.New(ConditionType("date"))
)

type Award struct {
	Type  AwardType `json:"type"`
	Value string    `json:"value"`
}

type Condition struct {
	Type  ConditionType `json:"type"`
	Op    string        `json:"op"`
	Value string        `json:"value"`
}

type Quest struct {
	Base

	ProjectID string
	Project   Project `gorm:"foreignKey:ProjectID"`

	Type           QuestType
	Status         QuestStatusType
	Index          int
	Title          string
	Description    string
	CategoryIDs    Array[string]
	Recurrence     RecurrenceType
	ValidationData string
	Awards         Array[Award]
	ConditionOp    ConditionOpType
	Conditions     Array[Condition]
}
