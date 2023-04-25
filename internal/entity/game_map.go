package entity

type GameMap struct {
	Base

	Name string `gorm:"unique"`

	// Init position.
	InitX int
	InitY int

	// Assets of game.
	Map    []byte
	Player []byte

	MapPath        string
	TileSetPath    string
	PlayerImgPath  string
	PlayerJSONPath string
}
