package gameengine

// topRight returns the top right position given the center, width, and height
// of an object.
func topRight(topLeft Position, s Size) Position {
	return Position{X: topLeft.X + s.Width, Y: topLeft.Y}
}

// bottomLeft returns the bottom left position given the center, width, and
// height of an object.
func bottomLeft(topLeft Position, s Size) Position {
	return Position{X: topLeft.X, Y: topLeft.Y + s.Height}
}

// bottomRight returns the bottom right position given the center, width, and
// height of an object.
func bottomRight(topLeft Position, s Size) Position {
	return Position{X: topLeft.X + s.Width, Y: topLeft.Y + s.Height}
}
