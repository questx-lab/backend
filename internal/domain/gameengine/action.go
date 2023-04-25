package gameengine

import (
	"errors"
	"time"

	"github.com/questx-lab/backend/internal/entity"
)

////////////////// MOVE Action
const maxMovingPixel = float64(20)
const minMovingPixel = float64(0.8)

type MoveAction struct {
	UserID    string
	Direction entity.DirectionType
	Position  Position
}

func (action *MoveAction) OnlyOwner() bool {
	return false
}

func (action *MoveAction) Apply(g *GameState) error {
	// Using map reverse to get the user position.
	user, ok := g.userMap[action.UserID]
	if !ok {
		return errors.New("invalid user id")
	}

	///////////////////// CHECK THE USER STATUS ////////////////////////////////

	if !user.IsActive {
		return errors.New("user not in room")
	}

	// Check if the user at the current position is standing on any collision
	// tile.
	if g.isObjectCollision(user.PixelPosition, playerWidth, playerHeight) {
		return errors.New("user is standing on a collision tile")
	}

	// The position client sends to server is the center of player, we need to
	// change it to a topleft position.
	newPosition := action.Position
	newPosition.X -= playerWidth / 2
	newPosition.Y -= playerHeight / 2

	// Check the distance between current and new position.
	d := user.PixelPosition.distance(newPosition)
	if d >= maxMovingPixel {
		return errors.New("move too fast")
	}

	if d <= minMovingPixel {
		return errors.New("move too slow")
	}

	// Check if the user at the new position is standing on any collision tile.
	if g.isObjectCollision(newPosition, playerWidth, playerHeight) {
		return errors.New("cannot go to a collision tile")
	}

	g.trackUserPosition(user.UserID, newPosition)

	return nil
}

////////////////// JOIN Action
type JoinAction struct {
	UserID string

	// These following fields is only assigned after applying into game state.
	position  Position
	direction entity.DirectionType
}

func (action *JoinAction) OnlyOwner() bool {
	return false
}

func (a *JoinAction) Apply(g *GameState) error {
	if user, ok := g.userMap[a.UserID]; ok {
		if user.IsActive {
			return errors.New("the user has already been active")
		}

		g.trackUserActive(a.UserID, true)
	} else {
		// Create a new user in game state with full information.
		g.addUser(User{
			UserID:        a.UserID,
			PixelPosition: Position{0, 0},
			Direction:     entity.Down,
			IsActive:      true,
			LastTimeMoved: time.Now(),
		})
	}

	// Update these fields to serialize to client.
	a.position = g.userMap[a.UserID].PixelPosition
	a.direction = g.userMap[a.UserID].Direction

	return nil
}

////////////////// EXIT Action
type ExitAction struct {
	UserID string
}

func (action *ExitAction) OnlyOwner() bool {
	return false
}

func (a *ExitAction) Apply(g *GameState) error {
	user, ok := g.userMap[a.UserID]
	if !ok {
		return errors.New("user has not appeared in room")
	}

	if !user.IsActive {
		return errors.New("the user is inactive, he must not have been appeared in game state")
	}

	g.trackUserActive(a.UserID, false)

	return nil
}

////////////////// INIT Action
// InitAction returns to the owner of this action the current game state.
type InitAction struct {
	UserID string

	initialUsers []User
}

func (action *InitAction) OnlyOwner() bool {
	return true
}

func (a *InitAction) Apply(g *GameState) error {
	serializedGameState := g.Serialize()
	a.initialUsers = serializedGameState.Users

	return nil
}
