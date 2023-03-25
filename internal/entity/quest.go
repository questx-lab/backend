package entity

import "github.com/questx-lab/backend/pkg/enum"

type QuestType string

var (
	QuestVisitLink = enum.New(QuestType("visit link"), "Visit Link")
	QuestText      = enum.New(QuestType("text"), "Text")
)

type QuestRecurrenceType string

var (
	QuestRecurrenceOnce    = enum.New(QuestRecurrenceType("once"), "Once")
	QuestRecurrenceDaily   = enum.New(QuestRecurrenceType("daily"), "Daily")
	QuestRecurrenceWeekly  = enum.New(QuestRecurrenceType("weekly"), "Weekly")
	QuestRecurrenceMonthly = enum.New(QuestRecurrenceType("monthly"), "Monthly")
)

type QuestStatusType string

var (
	QuestStatusDraft     = enum.New(QuestStatusType("draft"), "Draft")
	QuestStatusPublished = enum.New(QuestStatusType("published"), "Published")
	QuestStatusArchived  = enum.New(QuestStatusType("archived"), "Achived")
)

type QuestConditionOpType string

var (
	QuestConditionOpOr  = enum.New(QuestConditionOpType("or"), "OR")
	QuestConditionOpAnd = enum.New(QuestConditionOpType("and"), "AND")
)

type Quest struct {
	Base

	ProjectID string
	Project   Project `gorm:"foreignKey:ProjectID"`

	Type           QuestType
	Index          int
	Title          string
	Description    string
	CategoryIDs    string
	Recurrence     QuestRecurrenceType
	Status         QuestStatusType
	ValidationData string
	Awards         string
	ConditionOp    QuestConditionOpType
	Conditions     string
}
