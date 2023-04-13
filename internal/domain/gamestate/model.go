package gamestate

import (
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/entity"
)

type User struct {
	UserID   string
	ObjectID string

	// If the user presses the moving button which is the same with user's
	// direction, the game state treats it as a moving action.
	//
	// If the user presses a moving button which is difference from user's
	// direction, the game state treats it as rotating action.
	Direction entity.DirectionType

	// LastTimeMoved is the last time user uses the Moving Action. This is used
	// to track the moving speed of user.
	//
	// For example, if the last time moved of user is 10h12m9s, then the next
	// time user can move is 10h12m10s.
	LastTimeMoved time.Time
}

type Position struct {
	X int
	Y int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.X, p.Y)
}

func (p Position) move(direction entity.DirectionType) Position {
	switch direction {
	case entity.Up:
		return Position{X: p.X, Y: p.Y - 1}
	case entity.Down:
		return Position{X: p.X, Y: p.Y + 1}
	case entity.Right:
		return Position{X: p.X + 1, Y: p.Y}
	case entity.Left:
		return Position{X: p.X - 1, Y: p.Y}
	}

	return p
}

func (p Position) pixelToCell() Position {
	return Position{X: p.X/CellSize + 1, Y: p.Y/CellSize + 1}
}
