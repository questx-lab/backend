package entity

import (
	"time"

	"github.com/questx-lab/backend/pkg/enum"
	"gorm.io/gorm"
)

type UserAggregateRange string

var (
	UserAggregateRangeWeek  = enum.New(UserAggregateRange("week"))
	UserAggregateRangeMonth = enum.New(UserAggregateRange("month"))
	UserAggregateRangeTotal = enum.New(UserAggregateRange("total"))
)

var UserAggregateRangeList = []UserAggregateRange{
	UserAggregateRangeWeek,
	UserAggregateRangeMonth,
	UserAggregateRangeTotal,
}

type UserAggregate struct {
	ProjectID  string `gorm:"primaryKey"`
	UserID     string `gorm:"primaryKey"`
	RangeValue string `gorm:"primaryKey"`
	Range      UserAggregateRange

	TotalTask  uint64
	TotalPoint uint64

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" `
}
