package entity

import (
	"time"
)

type Migration struct {
	Version   int `gorm:"primaryKey"`
	CreatedAt time.Time
}
