package entity

type GameMapTileset struct {
	Base
	GameMapID  string
	GameMap    GameMap `gorm:"foreignKey:GameMapID"`
	TilesetURL string
}
