package entity

type GameMapPlayer struct {
	GameMapID string  `gorm:"primaryKey"`
	GameMap   GameMap `gorm:"foreignKey:GameMapID"`

	Name string `gorm:"primaryKey"`

	ImagePath string
	JSONPath  string
}
