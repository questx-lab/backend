package gamestate

// topLeft returns the top left position given the center, width, and height of
// an object.
func topLeft(center Position, width, height int) Position {
	return Position{X: center.X - width/2, Y: center.Y - height/2}
}

// topRight returns the top right position given the center, width, and height
// of an object.
func topRight(center Position, width, height int) Position {
	return Position{X: center.X + width/2, Y: center.Y - height/2}
}

// bottomLeft returns the bottom left position given the center, width, and
// height of an object.
func bottomLeft(center Position, width, height int) Position {
	return Position{X: center.X - width/2, Y: center.Y + height/2}
}

// bottomRight returns the bottom right position given the center, width, and
// height of an object.
func bottomRight(center Position, width, height int) Position {
	return Position{X: center.X + width/2, Y: center.Y + height/2}
}
