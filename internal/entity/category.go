package entity

type Category struct {
	Base
	Name      string
	ProjectID string  `gorm:"not null"`
	Project   Project `gorm:"foreignKey:ProjectID"`
	CreatedBy string  `gorm:"not null"`
}
