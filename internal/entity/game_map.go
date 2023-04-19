package entity

type GameMap struct {
	Base

	// Content is a serialized string of TMX file to display game map at client.
	Content string
}
