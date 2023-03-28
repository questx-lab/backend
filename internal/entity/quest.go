package entity

import (
	"github.com/questx-lab/backend/pkg/enum"
)

type QuestType string

var (
	QuestVisitLink = enum.New(QuestType("visit link"))
	QuestText      = enum.New(QuestType("text"))
)

type QuestRecurrenceType string

var (
	QuestRecurrenceOnce    = enum.New(QuestRecurrenceType("once"))
	QuestRecurrenceDaily   = enum.New(QuestRecurrenceType("daily"))
	QuestRecurrenceWeekly  = enum.New(QuestRecurrenceType("weekly"))
	QuestRecurrenceMonthly = enum.New(QuestRecurrenceType("monthly"))
)

type QuestStatusType string

var (
	QuestStatusDraft     = enum.New(QuestStatusType("draft"))
	QuestStatusPublished = enum.New(QuestStatusType("published"))
	QuestStatusArchived  = enum.New(QuestStatusType("archived"))
)

type QuestConditionOpType string

var (
	QuestConditionOpOr  = enum.New(QuestConditionOpType("or"))
	QuestConditionOpAnd = enum.New(QuestConditionOpType("and"))
)

type Award struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Condition struct {
	Type  string `json:"type"`
	Op    string `json:"op"`
	Value string `json:"value"`
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
	Recurrence     QuestRecurrenceType
	ValidationData string
	Awards         Array[Award]
	ConditionOp    QuestConditionOpType
	Conditions     Array[Condition]
}
