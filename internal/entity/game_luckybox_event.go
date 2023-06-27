package entity

import (
	"time"
)

type GameLuckyboxEvent struct {
	Base

	RoomID string
	Room   GameRoom `gorm:"foreignKey:RoomID"`

	Amount      int
	PointPerBox int
	IsRandom    bool
	StartTime   time.Time
	EndTime     time.Time
	IsStarted   bool
	IsStopped   bool
}
