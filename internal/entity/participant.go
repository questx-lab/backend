package entity

import (
	"database/sql"
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

	InviteCode    string `gorm:"unique"`
	InviteCount   uint64
	InvitedBy     sql.NullString
	InvitedByUser User `gorm:"foreignKey:InvitedBy"`
}
