package entity

type GameMap struct {
	Base

	Name string `gorm:"unique"`

	// Assets of game.
	Map []byte

	MapPath        string
	TileSetPath    string
	PlayerImgPath  string
	PlayerJSONPath string
}
