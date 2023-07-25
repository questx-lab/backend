package entity

type GameMap struct {
	Base
	Name      string `gorm:"unique"`
	ConfigURL string
}
