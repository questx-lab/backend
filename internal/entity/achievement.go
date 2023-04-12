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
	ProjectID string `gorm:"primaryKey"`
	UserID    string `gorm:"primaryKey"`
	// week/year or month/year
	Value string `gorm:"primaryKey"`
	Range AchievementRange

	TotalTask  uint64
	TotalPoint uint64

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
