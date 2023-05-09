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

var UserAggregateRangeList = []UserAggregateRange{
	UserAggregateRangeWeek,
	UserAggregateRangeMonth,
	UserAggregateRangeTotal,
}

type UserAggregate struct {
	ProjectID string `gorm:"primaryKey" json:"project_id"`
	UserID    string `gorm:"primaryKey" json:"user_id"`
	// week/year or month/year
	RangeValue string             `gorm:"primaryKey" json:"range_value"`
	Range      UserAggregateRange `json:"range"`

	TotalTask  uint64 `json:"total_task"`
	TotalPoint uint64 `json:"total_point"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
