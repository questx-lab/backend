package entity

import (
	"database/sql"
	"time"
)

type GameLuckyboxEvent struct {
	Base

	RoomID string
	Room   GameRoom `gorm:"foreignKey:RoomID"`

	Amount      int
	PointPerBox int
	StartTime   time.Time
	EndTime     sql.NullTime
	IsStarted   bool
	IsStopped   bool
}
