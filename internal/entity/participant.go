package entity

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type Follower struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	UserID string `gorm:"primaryKey"`
	User   User   `gorm:"foreignKey:UserID"`

	CommunityID string    `gorm:"primaryKey"`
	Community   Community `gorm:"foreignKey:CommunityID"`

	Points uint64
	Streak uint64

	InviteCode    string `gorm:"unique"`
	InviteCount   uint64
	InvitedBy     sql.NullString
	InvitedByUser User `gorm:"foreignKey:InvitedBy"`
}
