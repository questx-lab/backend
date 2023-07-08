package gameengine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/xcontext"
)

////////////////// MOVE Action
// TODO: Currently, we will disable checking max moving distance.
// const maxMovingPixel = float64(20)
const minMovingPixel = float64(5)

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

func (a *MoveAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	// Using map reverse to get the user position.
	user, ok := g.userMap[a.UserID]
	if !ok {
		return nil, errors.New("invalid user id")
	}

	if !user.ConnectedBy.Valid {
		return nil, errors.New("user not in room")
	}

	// Check if the user at the current position is standing on any collision
	// tile.
	if g.mapConfig.IsCollision(user.PixelPosition, user.Character.Size) {
		return nil, errors.New("user is standing on a collision tile")
	}

	// The position client sends to server is the center of character, we need
	// to change it to a topleft position.
	newPosition := a.Position.CenterToTopLeft(user.Character.Size)

	// Check the distance between the current position and the new one. If the
	// user is rotating, no need to check min distance.
	d := user.PixelPosition.Distance(newPosition)
	// TODO: Currently, we will disable checking max moving distance.
	// if d >= maxMovingPixel {
	// 	return errors.New("move too fast")
	// }
	if user.Direction == a.Direction && d <= minMovingPixel {
		return nil, errors.New("move too slow")
	}

	// Check if the user at the new position is standing on any collision tile.
	if g.mapConfig.IsCollision(newPosition, user.Character.Size) {
		return nil, errors.New("cannot go to a collision tile")
	}

	g.trackUserPosition(user.User.ID, a.Direction, newPosition)

	return nil, nil
}

////////////////// JOIN Action
type JoinAction struct {
	UserID string

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

func (a *JoinAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	if user, ok := g.userMap[a.UserID]; ok {
		if user.ConnectedBy.Valid {
			return nil, errors.New("the user has already been active")
		}

		g.trackUserPosition(a.UserID, entity.Down, g.initCenterPixelPosition.CenterToTopLeft(user.Character.Size))
		g.trackUserProxy(a.UserID, proxyID)
	} else {
		user, err := g.userRepo.GetByID(ctx, a.UserID)
		if err != nil {
			return nil, err
		}

		// By default, we  will choose the first one in our list.
		character := g.characters[0]
		if g.mapConfig.IsCollision(g.initCenterPixelPosition.CenterToTopLeft(character.Size), character.Size) {
			return nil, fmt.Errorf("init position %s is in collision with another object", character.Name)
		}

		// Create a new user in game state with full information.
		g.addUser(User{
			User: UserInfo{
				ID:        user.ID,
				Name:      user.Name,
				AvatarURL: user.ProfilePicture,
			},
			Character:      character,
			PixelPosition:  g.initCenterPixelPosition.CenterToTopLeft(character.Size),
			Direction:      entity.Down,
			ConnectedBy:    sql.NullString{Valid: true, String: proxyID},
			LastTimeAction: make(map[string]time.Time),
		})
	}

	// Update these fields to serialize to client.
	a.user = *g.userMap[a.UserID]

	return nil, nil
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

func (a *ExitAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	user, ok := g.userMap[a.UserID]
	if !ok {
		return nil, errors.New("user has not appeared in room")
	}

	if !user.ConnectedBy.Valid {
		return nil, errors.New("the user is inactive, he must not have been appeared in game state")
	}

	g.trackUserProxy(a.UserID, "")
	// TODO: This action will reset the position after user exits room.
	// The is using for testing with frontend. If the frontend completed, MUST
	// remove this code.
	g.trackUserPosition(a.UserID, entity.Down, g.initCenterPixelPosition.CenterToTopLeft(user.Character.Size))

	return nil, nil
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

func (a *InitAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	user, ok := g.userMap[a.UserID]
	if !ok || !user.ConnectedBy.Valid {
		return nil, errors.New("user is not in map")
	}

	a.initialUsers = g.Serialize()
	a.messageHistory = g.messageHistory
	for i := range g.luckyboxes {
		a.luckyboxes = append(a.luckyboxes,
			g.luckyboxes[i].WithCenterPixelPosition(g.mapConfig.TileSizeInPixel))
	}

	return nil, nil
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

func (a *MessageAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	user, ok := g.userMap[a.UserID]
	if !ok || !user.ConnectedBy.Valid {
		return nil, errors.New("user is not in map")
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

	return nil, nil
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

func (a *EmojiAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	user, ok := g.userMap[a.UserID]
	if !ok || !user.ConnectedBy.Valid {
		return nil, errors.New("user is not in map")
	}

	return nil, nil
}

////////////////// START LUCKYBOX EVENT Action
// StartLuckyboxEventAction generates luckybox in room.
type StartLuckyboxEventAction struct {
	UserID  string
	EventID string

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

func (a *StartLuckyboxEventAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	if a.UserID != "" {
		// Regular user cannot send create_luckybox_event action.
		// Only our service can trigger this action.
		return nil, errors.New("permission denied")
	}

	event, err := g.gameRepo.GetLuckyboxEventByID(ctx, a.EventID)
	if err != nil {
		return nil, err
	}

	createdBoxes := 0
	retry := 0
	for createdBoxes < event.Amount && retry < xcontext.Configs(ctx).Game.LuckyboxGenerateMaxRetry {
		tilePosition := Position{
			X: crypto.RandIntn(g.mapConfig.MapSizeInTile.Width),
			Y: crypto.RandIntn(g.mapConfig.MapSizeInTile.Height),
		}

		if _, ok := g.mapConfig.ReachableTileMap[tilePosition]; !ok {
			retry++
			continue
		}

		if _, ok := g.luckyboxesByTilePosition[tilePosition]; ok {
			retry++
			continue
		}

		point := event.PointPerBox
		if event.IsRandom {
			point = crypto.RandRange(1, event.PointPerBox+1)
		}

		luckybox := Luckybox{
			ID:            uuid.NewString(),
			EventID:       a.EventID,
			Point:         point,
			PixelPosition: g.mapConfig.tileToPixel(tilePosition),
		}

		g.addLuckybox(luckybox)
		a.newLuckyboxes = append(a.newLuckyboxes, luckybox.WithCenterPixelPosition(g.mapConfig.MapSizeInTile))

		createdBoxes++
		retry = 0
	}

	return nil, nil
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

func (a *StopLuckyboxEventAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	if a.UserID != "" {
		// Regular user cannot send stop_luckybox_event action.
		// Only our service can trigger this action.
		return nil, errors.New("permission denied")
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

	return nil, nil
}

////////////////// COLLECT LUCKYBOX Action
// CollectLuckyboxAction is used to user collect the luckybox.
// TODO: Need to determine the exact value of the following value in frontend.
const collectMinTileDistance = float64(2)

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

func (a *CollectLuckyboxAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	user, ok := g.userMap[a.UserID]
	if !ok {
		return nil, errors.New("user is not in map")
	}

	luckybox, ok := g.luckyboxes[a.LuckyboxID]
	if !ok {
		return nil, errors.New("luckybox doesn't exist")
	}

	userTilePosition := g.mapConfig.pixelToTile(user.PixelPosition)
	luckyboxTilePosition := g.mapConfig.pixelToTile(luckybox.PixelPosition)
	if userTilePosition.Distance(luckyboxTilePosition) > collectMinTileDistance {
		return nil, errors.New("too far to collect luckybox")
	}

	err := g.followerRepo.IncreasePoint(ctx, a.UserID, g.communityID, uint64(luckybox.Point), false)
	if err != nil {
		return nil, err
	}

	err = g.leaderboard.ChangePointLeaderboard(ctx, int64(luckybox.Point), time.Now(),
		a.UserID, g.communityID)
	if err != nil {
		return nil, err
	}

	g.removeLuckybox(luckybox.ID, a.UserID)
	a.luckybox = luckybox.WithCenterPixelPosition(g.mapConfig.TileSizeInPixel)

	return nil, nil
}

////////////////// CLEANUP PROXY Action
// CleanupProxyAction is used to clean up disconnected proxy.
type CleanupProxyAction struct {
	UserID       string
	LiveProxyIDs []string
}

func (a CleanupProxyAction) SendTo() []string {
	// Send to noone.
	return []string{}
}

func (a CleanupProxyAction) Type() string {
	return "cleanup_proxy"
}

func (a CleanupProxyAction) Owner() string {
	return ""
}

func (a *CleanupProxyAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	liveProxyMap := map[string]any{}
	for _, proxyID := range a.LiveProxyIDs {
		liveProxyMap[proxyID] = nil
	}

	exitActions := []model.GameActionServerRequest{}

	for _, user := range g.userMap {
		if user.ConnectedBy.Valid {
			if _, ok := liveProxyMap[user.ConnectedBy.String]; !ok {
				exitActions = append(exitActions, model.GameActionServerRequest{
					UserID: user.User.ID,
					Type:   ExitAction{}.Type(),
				})
			}
		}
	}

	return exitActions, nil
}
