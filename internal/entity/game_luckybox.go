package entity

import "database/sql"

type GameLuckybox struct {
	Base

	EventID string
	Event   GameLuckyboxEvent `gorm:"foreignKey:EventID"`

	PositionX       int
	PositionY       int
	Point           int
	IsRandom        bool
	CollectedBy     sql.NullString
	CollectedByUser User `gorm:"foreignKey:CollectedBy"`
}
