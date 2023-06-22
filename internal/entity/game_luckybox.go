package entity

type GameLuckybox struct {
	Base

	EventID string
	Event   GameLuckyboxEvent `gorm:"foreignKey:EventID"`

	PositionX   int
	PositionY   int
	Point       int
	IsCollected bool
}
