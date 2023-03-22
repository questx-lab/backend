package entity

import "time"

type User struct {
	ID      string `gorm:"primaryKey"`
	Address string
	Name    string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}
