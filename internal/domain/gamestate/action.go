package gamestate

import (
	"errors"
	"time"

	"github.com/questx-lab/backend/internal/entity"
)

const movingDelay = 100 * time.Millisecond

var oppositeDirection = map[entity.DirectionType]entity.DirectionType{
	entity.Up:    entity.Down,
	entity.Down:  entity.Up,
	entity.Left:  entity.Right,
	entity.Right: entity.Left,
}

type UserMoveAction struct {
	UserID    string
	Direction entity.DirectionType
}

func NewUserMoveAction(userID string, direction entity.DirectionType) *UserMoveAction {
	return &UserMoveAction{UserID: userID, Direction: direction}
}

func (action *UserMoveAction) Apply(g *GameState) error {
	// Using map reverse to get the user position.
	userPixelPosition, ok := g.userMapInverse[action.UserID]
	if !ok {
		return errors.New("invalid user id")
	}

	///////////////////// CHECK THE USER STATUS ////////////////////////////////

	// Check if the user is actually at this position on the map.
	user, ok := g.userMap[userPixelPosition]
	if !ok {
		return errors.New("game state is wrong")
	}

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
		g.userMap[userPixelPosition] = user

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
	if g.isObjectCollision(userPixelPosition, playerWidth, playerHeight) {
		return errors.New("user is standing on a collision tile")
	}

	// Get the new position after moving and check if the new position is valid.
	newUserPixelPosition := userPixelPosition.move(action.Direction)
	if newUserPixelPosition == userPixelPosition {
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
	g.userMap[newUserPixelPosition] = user
	g.userMapInverse[user.UserID] = newUserPixelPosition

	// Remove user at the old position.
	delete(g.userMap, userPixelPosition)

	// Keep track database difference. The direction doesn't change.
	g.updateUserDiff(entity.GameUser{
		UserID:    user.UserID,
		PositionX: newUserPixelPosition.X,
		PositionY: newUserPixelPosition.Y,
	})

	return nil
}
