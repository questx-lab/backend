package gamestate

import (
	"errors"
	"time"

	"github.com/questx-lab/backend/internal/entity"
)

////////////////// MOVE Action
const movingDelay = 100 * time.Millisecond

type MoveAction struct {
	UserID    string
	Direction entity.DirectionType
}

func (action *MoveAction) Apply(g *GameState) error {
	// Using map reverse to get the user position.
	user, ok := g.userMap[action.UserID]
	if !ok {
		return errors.New("invalid user id")
	}

	///////////////////// CHECK THE USER STATUS ////////////////////////////////

	// Check the current direction of user. If the current direction of user is
	// difference from the direction of action, this action will become a
	// rotating action.
	if user.Direction != action.Direction {
		g.trackUserDirection(user.UserID, action.Direction)
		return nil
	}

	// Check the moving delay.
	if time.Since(user.LastTimeMoved) < movingDelay {
		return errors.New("move too fast")
	}

	///////////////////// CHECK THE CURRENT POSITION ///////////////////////////

	// Check if the user at the current position is standing on any collision
	// tile.
	if g.isObjectCollision(user.PixelPosition, playerWidth, playerHeight) {
		return errors.New("user is standing on a collision tile")
	}

	///////////////////// CHECK THE NEW POSITION ///////////////////////////////

	// Get the new position after moving.
	newUserPixelPosition := user.PixelPosition.move(action.Direction)
	if newUserPixelPosition == user.PixelPosition {
		return errors.New("not change position after moving")
	}

	// Check if the user at the new position is standing on any collision tile.
	if g.isObjectCollision(newUserPixelPosition, playerWidth, playerHeight) {
		return errors.New("cannot go to a collision tile")
	}

	g.trackUserPosition(user.UserID, newUserPixelPosition)

	return nil
}

////////////////// JOIN Action
type JoinAction struct {
	UserID string

	// These following fields is only assigned after applying into game state.
	position  Position
	direction entity.DirectionType
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
