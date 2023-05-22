package entity

import "database/sql"

type Badge struct {
	UserID      string         `gorm:"primaryKey"`
	User        User           `gorm:"foreignKey:UserID"`
	ProjectID   sql.NullString `gorm:"primaryKey"`
	Project     Project        `gorm:"foreignKey:ProjectID"`
	Name        string         `gorm:"primaryKey"`
	Level       int
	WasNotified bool
}
