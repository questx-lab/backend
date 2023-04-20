package entity

type GameMap struct {
	Base

	Name string `gorm:"unique"`

	// Content is a serialized string of TMX file to display game map at client.
	Content []byte
}
