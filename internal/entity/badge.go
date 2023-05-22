package entity

import "database/sql"

type Badge struct {
	UserID      string         `gorm:"primaryKey"`
	User        User           `gorm:"foreignKey:UserID"`
	CommunityID sql.NullString `gorm:"primaryKey"`
	Community   Community      `gorm:"foreignKey:CommunityID"`
	Name        string         `gorm:"primaryKey"`
	Level       int
	WasNotified bool
}
