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

	ReferralCode  string `gorm:"unique"`
	ReferralCount uint64
	ReferralID    sql.NullString
	Referral      User `gorm:"foreignKey:ReferralID"`
}
