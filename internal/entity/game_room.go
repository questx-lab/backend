package entity

type GameRoom struct {
	Base
	Name    string
	MapID   string
	GameMap GameMap `gorm:"foreignKey:MapID"`
}
