package entity

type GameMapPlayer struct {
	Base

	GameMapID string  `gorm:"index:idx_map_id_name,unique"`
	GameMap   GameMap `gorm:"foreignKey:GameMapID"`

	Name string `gorm:"index:idx_map_id_name,unique"`

	ConfigURL string
	ImageURL  string
}
