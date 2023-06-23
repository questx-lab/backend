package gameengine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/xcontext"
)

////////////////// MOVE Action
// TODO: Currently, we will disable checking max moving distance.
// const maxMovingPixel = float64(20)
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
	newPosition := a.Position.CenterToTopLeft(user.Player.Size)

	// Check the distance between the current position and the new one. If the
	// user is rotating, no need to check min distance.
	d := user.PixelPosition.Distance(newPosition)
	// TODO: Currently, we will disable checking max moving distance.
	// if d >= maxMovingPixel {
	// 	return errors.New("move too fast")
	// }
	if user.Direction == a.Direction && d <= minMovingPixel {
		return errors.New("move too slow")
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

		if g.mapConfig.IsPlayerCollision(g.initCenterPos.CenterToTopLeft(player.Size), player) {
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
			PixelPosition:  g.initCenterPos.CenterToTopLeft(player.Size),
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
	// TODO: This action will reset the position after user exits room.
	// The is using for testing with frontend. If the frontend completed, MUST
	// remove this code.
	g.trackUserPosition(a.UserID, entity.Down, g.initCenterPos.CenterToTopLeft(user.Player.Size))

	return nil
}

////////////////// INIT Action
// InitAction returns to the owner of this action the current game state.
type InitAction struct {
	UserID string

	initialUsers   []User
	messageHistory []Message
	luckyboxes     []Luckybox
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
	for i := range g.luckyboxes {
		a.luckyboxes = append(a.luckyboxes,
			g.luckyboxes[i].WithCenterPixelPosition(g.mapConfig.TileSizeInPixel))
	}

	return nil
}

////////////////// MESSAGE Action
// MessageAction sends message to game.
type MessageAction struct {
	UserID    string
	Message   string
	CreatedAt time.Time

	user UserInfo
}

func (a MessageAction) SendTo() []string {
	// Send to everyone.
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
		User:      user.User,
		Message:   a.Message,
		CreatedAt: a.CreatedAt,
	})

	a.user = user.User

	return nil
}

////////////////// EMOJI Action
// EmojiAction sends emoji to game.
type EmojiAction struct {
	UserID string
	Emoji  string
}

func (a EmojiAction) SendTo() []string {
	// Send to every one.
	return nil
}

func (a EmojiAction) Type() string {
	return "emoji"
}

func (a EmojiAction) Owner() string {
	return a.UserID
}

func (a *EmojiAction) Apply(ctx context.Context, g *GameState) error {
	user, ok := g.userMap[a.UserID]
	if !ok || !user.IsActive {
		return errors.New("user is not in map")
	}

	return nil
}

////////////////// START LUCKYBOX EVENT Action
// StartLuckyboxEventAction generates luckybox in room.
type StartLuckyboxEventAction struct {
	UserID      string
	EventID     string
	Amount      int
	PointPerBox int

	newLuckyboxes []Luckybox
}

func (a StartLuckyboxEventAction) SendTo() []string {
	// Send to everyone.
	return nil
}

func (a StartLuckyboxEventAction) Type() string {
	return "start_luckybox_event"
}

func (a StartLuckyboxEventAction) Owner() string {
	// This action not belongs to any user. Our service triggers it.
	return ""
}

func (a *StartLuckyboxEventAction) Apply(ctx context.Context, g *GameState) error {
	if a.UserID != "" {
		// Regular user cannot send create_luckybox_event action.
		// Only our service can trigger this action.
		return errors.New("permission denied")
	}

	createdBoxes := 0
	retry := 0
	for createdBoxes < a.Amount && retry < xcontext.Configs(ctx).Game.LuckyboxGenerateMaxRetry {
		tilePosition := Position{
			X: crypto.RandIntn(g.mapConfig.MapSizeInTile.Width),
			Y: crypto.RandIntn(g.mapConfig.MapSizeInTile.Height),
		}
		if _, ok := g.mapConfig.CollisionTileMap[tilePosition]; ok {
			retry++
			continue
		}

		if _, ok := g.luckyboxesByTilePosition[tilePosition]; ok {
			retry++
			continue
		}

		luckybox := Luckybox{
			ID:            uuid.NewString(),
			EventID:       a.EventID,
			Point:         a.PointPerBox,
			PixelPosition: g.mapConfig.tileToPixel(tilePosition),
		}

		g.addLuckybox(luckybox)
		a.newLuckyboxes = append(a.newLuckyboxes, luckybox.WithCenterPixelPosition(g.mapConfig.MapSizeInTile))

		createdBoxes++
		retry = 0
	}

	return nil
}

////////////////// STOP LUCKYBOX EVENT Action
// StopLuckyboxEventAction generates luckybox in room.
type StopLuckyboxEventAction struct {
	UserID  string
	EventID string

	removedLuckyboxes []Luckybox
}

func (a StopLuckyboxEventAction) SendTo() []string {
	// Send to everyone.
	return nil
}

func (a StopLuckyboxEventAction) Type() string {
	return "stop_luckybox_event"
}

func (a StopLuckyboxEventAction) Owner() string {
	// This action not belongs to any user. Our service triggers it.
	return ""
}

func (a *StopLuckyboxEventAction) Apply(ctx context.Context, g *GameState) error {
	if a.UserID != "" {
		// Regular user cannot send stop_luckybox_event action.
		// Only our service can trigger this action.
		return errors.New("permission denied")
	}

	for _, luckybox := range g.luckyboxes {
		if luckybox.EventID == a.EventID {
			a.removedLuckyboxes = append(a.removedLuckyboxes,
				luckybox.WithCenterPixelPosition(g.mapConfig.TileSizeInPixel))
		}
	}

	for _, luckybox := range a.removedLuckyboxes {
		g.removeLuckybox(luckybox.ID, "")
	}

	return nil
}

////////////////// COLLECT LUCKYBOX Action
// CollectLuckyboxAction is used to user collect the luckybox.
// TODO: Need to determine the exact value of the following value in frontend.
const collect_min_tile_distance = float64(2)

type CollectLuckyboxAction struct {
	UserID     string
	LuckyboxID string

	luckybox Luckybox
}

func (a CollectLuckyboxAction) SendTo() []string {
	// Send to everyone.
	return nil
}

func (a CollectLuckyboxAction) Type() string {
	return "collect_luckybox"
}

func (a CollectLuckyboxAction) Owner() string {
	return a.UserID
}

func (a *CollectLuckyboxAction) Apply(ctx context.Context, g *GameState) error {
	user, ok := g.userMap[a.UserID]
	if !ok {
		return errors.New("user is not in map")
	}

	luckybox, ok := g.luckyboxes[a.LuckyboxID]
	if !ok {
		return errors.New("luckybox doesn't exist")
	}

	userTilePosition := g.mapConfig.pixelToTile(user.PixelPosition)
	luckyboxTilePosition := g.mapConfig.pixelToTile(luckybox.PixelPosition)
	if userTilePosition.Distance(luckyboxTilePosition) > collect_min_tile_distance {
		return errors.New("too far to collect luckybox")
	}

	err := g.followerRepo.IncreasePoint(ctx, a.UserID, g.communityID, uint64(luckybox.Point), false)
	if err != nil {
		return err
	}

	err = g.leaderboard.ChangePointLeaderboard(ctx, int64(luckybox.Point), time.Now(),
		a.UserID, g.communityID)
	if err != nil {
		return err
	}

	g.removeLuckybox(luckybox.ID, a.UserID)
	a.luckybox = luckybox.WithCenterPixelPosition(g.mapConfig.TileSizeInPixel)

	return nil
}
