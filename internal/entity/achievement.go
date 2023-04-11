package entity

import (
	"time"

	"gorm.io/gorm"
)

type AchievementRange string

const (
	AchievementRangeWeek  AchievementRange = "week"
	AchievementRangeMonth AchievementRange = "month"
	AchievementRangeTotal AchievementRange = "total"
)

var AchievementRangeList = []AchievementRange{AchievementRangeWeek, AchievementRangeMonth, AchievementRangeTotal}

type Achievement struct {
	ProjectID string
	UserID    string
	Range     AchievementRange

	TotalTask int
	TotalExp  int64

	// week/year or month/year
	Value string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
