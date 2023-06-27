package entity

import (
	"database/sql"
	"time"
)

type BadgeDetail struct {
	UserID      string         `gorm:"primaryKey"`
	User        User           `gorm:"foreignKey:UserID"`
	CommunityID sql.NullString `gorm:"primaryKey"`
	Community   Community      `gorm:"foreignKey:CommunityID"`
	BadgeID     string         `gorm:"primaryKey"`
	Badge       Badge          `gorm:"foreignKey:BadgeID"`
	WasNotified bool
	CreatedAt   time.Time
}
