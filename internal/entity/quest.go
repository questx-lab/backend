package entity

type Quest struct {
	Base

	ProjectID string
	Project   Project `gorm:"foreignKey:ProjectID"`

	Type           string
	Index          int
	Title          string
	Description    string
	CategoryIDs    string
	Recurrence     string
	Status         string
	ValidationData string
	Awards         string
	ConditionOp    string
	Conditions     string
}
