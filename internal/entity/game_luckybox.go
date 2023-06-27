package entity

import "database/sql"

type GameLuckybox struct {
	Base

	EventID string
	Event   GameLuckyboxEvent `gorm:"foreignKey:EventID"`

	PositionX       int
	PositionY       int
	Point           int
	CollectedBy     sql.NullString
	CollectedByUser User `gorm:"foreignKey:CollectedBy"`
	CollectedAt     sql.NullTime
}
