package entity

type GameRoom struct {
	Base
	CommunityID string
	Community   Community `gorm:"foreignKey:CommunityID"`
	MapID       string
	GameMap     GameMap `gorm:"foreignKey:MapID"`
	Name        string
	StartedBy   string
}
