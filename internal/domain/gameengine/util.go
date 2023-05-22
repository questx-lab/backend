package gameengine

// topRight returns the top right position given the center, width, and height
// of an object.
func topRight(topLeft Position, width, height int) Position {
	return Position{X: topLeft.X + width, Y: topLeft.Y}
}

// bottomLeft returns the bottom left position given the center, width, and
// height of an object.
func bottomLeft(topLeft Position, width, height int) Position {
	return Position{X: topLeft.X, Y: topLeft.Y + height}
}

// bottomRight returns the bottom right position given the center, width, and
// height of an object.
func bottomRight(topLeft Position, width, height int) Position {
	return Position{X: topLeft.X + width, Y: topLeft.Y + height}
}
