package entity

type GameBlockedCell struct {
	Base

	MapID   string
	GameMap GameMap `gorm:"foreignKey:MapID"`

	PositionX int
	PositionY int
}
