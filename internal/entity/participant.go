package entity

import (
	"time"

	"gorm.io/gorm"
)

type Participant struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	UserID string `gorm:"primaryKey"`
	User   User   `gorm:"foreignKey:UserID"`

	ProjectID string  `gorm:"primaryKey"`
	Project   Project `gorm:"foreignKey:ProjectID"`

	Points uint64
}
