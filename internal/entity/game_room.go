package entity

type GameRoom struct {
	Base
	Name    string `gorm:"unique"`
	MapID   string
	GameMap GameMap `gorm:"foreignKey:MapID"`
}
