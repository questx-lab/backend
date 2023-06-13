package gameengine

import (
	"context"
	"errors"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

////////////////// MOVE Action
const maxMovingPixel = float64(20)
const minMovingPixel = float64(0.8)

type MoveAction struct {
	UserID    string
	Direction entity.DirectionType
	Position  Position
}

func (a MoveAction) SendTo() []string {
	// Broadcast to all clients.
	return nil
}

func (a MoveAction) Type() string {
	return "move"
}

func (a MoveAction) Owner() string {
	return a.UserID
}

func (a *MoveAction) Apply(ctx context.Context, g *GameState) error {
	// Using map reverse to get the user position.
	user, ok := g.userMap[a.UserID]
	if !ok {
		return errors.New("invalid user id")
	}

	if !user.IsActive {
		return errors.New("user not in room")
	}

	// Check if the user at the current position is standing on any collision
	// tile.
	if g.isObjectCollision(user.PixelPosition, g.playerWidth, g.playerHeight) {
		return errors.New("user is standing on a collision tile")
	}

	// The position client sends to server is the center of player, we need to
	// change it to a topleft position.
	newPosition := a.Position.centerToTopLeft(g.playerWidth, g.playerHeight)

	// Check the distance between current and new position.
	d := user.PixelPosition.distance(newPosition)
	if d >= maxMovingPixel {
		return errors.New("move too fast")
	}

	if d <= minMovingPixel {
		return errors.New("move too slow")
	}

	// Check if the user at the new position is standing on any collision tile.
	if g.isObjectCollision(newPosition, g.playerWidth, g.playerHeight) {
		return errors.New("cannot go to a collision tile")
	}

	g.trackUserPosition(user.UserID, a.Direction, newPosition)

	return nil
}

////////////////// JOIN Action
type JoinAction struct {
	UserID string

	// These following fields is only assigned after applying into game state.
	position  Position
	direction entity.DirectionType
}

func (a JoinAction) SendTo() []string {
	// Broadcast to all clients.
	return nil
}

func (a JoinAction) Type() string {
	return "join"
}

func (a JoinAction) Owner() string {
	return a.UserID
}

func (a *JoinAction) Apply(ctx context.Context, g *GameState) error {
	if user, ok := g.userMap[a.UserID]; ok {
		if user.IsActive {
			return errors.New("the user has already been active")
		}

		g.trackUserActive(a.UserID, true)
	} else {
		// Create a new user in game state with full information.
		g.addUser(User{
			UserID:         a.UserID,
			PixelPosition:  g.initialPosition,
			Direction:      entity.Down,
			IsActive:       true,
			LastTimeAction: make(map[string]time.Time),
		})
	}

	// Update these fields to serialize to client.
	a.position = g.userMap[a.UserID].PixelPosition.topLeftToCenter(g.playerWidth, g.playerHeight)
	a.direction = g.userMap[a.UserID].Direction

	return nil
}

////////////////// EXIT Action
type ExitAction struct {
	UserID string
}

func (a ExitAction) SendTo() []string {
	// Broadcast to all clients.
	return nil
}

func (a ExitAction) Type() string {
	return "exit"
}

func (a ExitAction) Owner() string {
	return a.UserID
}

func (a *ExitAction) Apply(ctx context.Context, g *GameState) error {
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

	initialUsers   []User
	messageHistory []Message
}

func (a InitAction) SendTo() []string {
	// Send to only the owner of action.
	return []string{a.UserID}
}

func (a InitAction) Type() string {
	return "init"
}

func (a InitAction) Owner() string {
	return a.UserID
}

func (a *InitAction) Apply(ctx context.Context, g *GameState) error {
	user, ok := g.userMap[a.UserID]
	if !ok || !user.IsActive {
		return errors.New("user is not in map")
	}

	a.initialUsers = g.Serialize()
	a.messageHistory = g.messageHistory
	return nil
}

////////////////// MESSAGE Action
// MessageAction sends message to game.

type MessageAction struct {
	UserID    string
	Message   string
	CreatedAt time.Time
}

func (a MessageAction) SendTo() []string {
	// Send to every one.
	return nil
}

func (a MessageAction) Type() string {
	return "message"
}

func (a MessageAction) Owner() string {
	return a.UserID
}

func (a *MessageAction) Apply(ctx context.Context, g *GameState) error {
	user, ok := g.userMap[a.UserID]
	if !ok || !user.IsActive {
		return errors.New("user is not in map")
	}

	if len(g.messageHistory) >= xcontext.Configs(ctx).Game.MessageHistoryLength {
		// Remove the oldest message from history.
		g.messageHistory = g.messageHistory[1:]
	}

	g.messageHistory = append(g.messageHistory, Message{
		UserID:    a.UserID,
		Message:   a.Message,
		CreatedAt: a.CreatedAt,
	})
	return nil
}
