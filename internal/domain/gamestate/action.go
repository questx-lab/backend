package gamestate

import (
	"errors"
	"time"

	"github.com/questx-lab/backend/internal/entity"
)

const movingDelay = 100 * time.Millisecond

type Move struct {
	UserID    string
	Direction entity.DirectionType
}

func NewMove(userID string, direction entity.DirectionType) *Move {
	return &Move{UserID: userID, Direction: direction}
}

func (action *Move) Apply(g *GameState) error {
	// Using map reverse to get the user position.
	user, ok := g.userMap[action.UserID]
	if !ok {
		return errors.New("invalid user id")
	}

	///////////////////// CHECK THE USER STATUS ////////////////////////////////

	// Check if the user is actually at this position on the map.

	// Check the moving delay.
	if time.Since(user.LastTimeMoved) < movingDelay {
		return errors.New("move too fast")
	}

	// Check the current direction of user. If the current direction of user is
	// difference from the direction of action, this action will become a
	// rotating action.
	// NOTE: No need to update the last time moved, because this is not a moving
	// action.
	if user.Direction != action.Direction {
		user.Direction = action.Direction
		g.userMap[action.UserID] = user

		// Keep track database difference. The position doesn't change.
		g.updateUserDiff(entity.GameUser{
			UserID:    user.UserID,
			Direction: user.Direction,
		})
		return nil
	}

	///////////////////// CHECK THE CURRENT POSITION ///////////////////////////

	// Check if the user at the current position is standing on any collision
	// tile.
	if g.isObjectCollision(user.PixelPosition, playerWidth, playerHeight) {
		return errors.New("user is standing on a collision tile")
	}

	// Get the new position after moving and check if the new position is valid.
	newUserPixelPosition := user.PixelPosition.move(action.Direction)
	if newUserPixelPosition == user.PixelPosition {
		return errors.New("not change position after moving")
	}

	newUserPosition := g.pixelToTile(newUserPixelPosition)
	if newUserPosition.X < 0 || newUserPosition.X >= g.height {
		return errors.New("invalid position x")
	}

	if newUserPosition.Y < 0 || newUserPosition.Y >= g.width {
		return errors.New("invalid position y")
	}

	///////////////////// CHECK THE NEW POSITION ///////////////////////////////

	// Check if the user at the new position is standing on any collision tile.
	if g.isObjectCollision(newUserPixelPosition, playerWidth, playerHeight) {
		return errors.New("cannot go to a non-existed tile")
	}

	// Move user to the new position.
	user.LastTimeMoved = time.Now()
	user.PixelPosition = newUserPixelPosition
	g.userMap[user.UserID] = user

	// Keep track database difference. The direction doesn't change.
	g.updateUserDiff(entity.GameUser{
		UserID:    user.UserID,
		PositionX: newUserPixelPosition.X,
		PositionY: newUserPixelPosition.Y,
	})

	return nil
}
