package entity

import (
	"time"

	"gorm.io/gorm"
)

type UserAggregateRange string

const (
	UserAggregateRangeWeek  UserAggregateRange = "week"
	UserAggregateRangeMonth UserAggregateRange = "month"
	UserAggregateRangeTotal UserAggregateRange = "total"
)

var UserAggregateRangeList = []UserAggregateRange{UserAggregateRangeWeek, UserAggregateRangeMonth, UserAggregateRangeTotal}

type UserAggregate struct {
	ProjectID string `gorm:"primaryKey"`
	UserID    string `gorm:"primaryKey"`
	// week/year or month/year
	Value string `gorm:"primaryKey"`
	Range UserAggregateRange

	TotalTask  uint64
	TotalPoint uint64

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
