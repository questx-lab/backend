package entity

type GameMap struct {
	Base

	Name string `gorm:"unique"`

	// Init position.
	InitX int
	InitY int

	ConfigURL       string
	CollisionLayers string
}
