package gameengine

import (
	"context"
	"errors"
	"fmt"
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
	if g.mapConfig.IsPlayerCollision(user.PixelPosition, user.Player) {
		return errors.New("user is standing on a collision tile")
	}

	// The position client sends to server is the center of player, we need to
	// change it to a topleft position.
	newPosition := a.Position.CenterToTopLeft(user.Player)

	// Check the distance between the current position and the new one. If the
	// user is rotating, no need to check.
	if user.Direction == a.Direction {
		d := user.PixelPosition.Distance(newPosition)
		if d >= maxMovingPixel {
			return errors.New("move too fast")
		}

		if d <= minMovingPixel {
			return errors.New("move too slow")
		}
	}

	// Check if the user at the new position is standing on any collision tile.
	if g.mapConfig.IsPlayerCollision(newPosition, user.Player) {
		return errors.New("cannot go to a collision tile")
	}

	g.trackUserPosition(user.User.ID, a.Direction, newPosition)

	return nil
}

////////////////// JOIN Action
type JoinAction struct {
	UserID string

	// User only need to specify this field if he never joined this room before.
	PlayerName string

	// These following fields is only assigned after applying into game state.
	user User
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
		user, err := g.userRepo.GetByID(ctx, a.UserID)
		if err != nil {
			return err
		}

		// By default, if user doesn't explicitly choose the player name, we
		// will choose the first one in our list.
		player := g.players[0]
		if a.PlayerName != "" {
			found := false
			for _, p := range g.players {
				if p.Name == a.PlayerName {
					found = true
					player = p
				}
			}

			if !found {
				return fmt.Errorf("not found player %s", a.PlayerName)
			}
		}

		if g.mapConfig.IsPlayerCollision(g.initCenterPos.CenterToTopLeft(player), player) {
			return fmt.Errorf("init position %s is in collision with another object", player.Name)
		}

		// Create a new user in game state with full information.
		g.addUser(User{
			User: UserInfo{
				ID:        user.ID,
				Name:      user.Name,
				AvatarURL: user.ProfilePicture,
			},
			Player:         player,
			PixelPosition:  g.initCenterPos.CenterToTopLeft(player),
			Direction:      entity.Down,
			IsActive:       true,
			LastTimeAction: make(map[string]time.Time),
		})
	}

	// Update these fields to serialize to client.
	a.user = *g.userMap[a.UserID]

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
