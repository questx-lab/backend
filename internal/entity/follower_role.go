package entity

import (
	"time"
)

type FollowerRole struct {
	CreatedAt time.Time

	UserID string `gorm:"primaryKey"`
	User   User   `gorm:"foreignKey:UserID"`

	CommunityID string    `gorm:"primaryKey"`
	Community   Community `gorm:"foreignKey:CommunityID"`

	RoleID string `gorm:"primaryKey"`
	Role   Role   `gorm:"foreignKey:RoleID"`
}
