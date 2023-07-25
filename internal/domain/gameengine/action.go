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
	OwnerID   string
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
	return a.OwnerID
}

func (a *MoveAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	// Using map reverse to get the user position.
	user, ok := g.userMap[a.OwnerID]
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
	OwnerID string

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
	return a.OwnerID
}

func (a *JoinAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	if user, ok := g.userMap[a.OwnerID]; ok {
		if user.ConnectedBy.Valid {
			return nil, errors.New("the user has already been active")
		}

		g.trackUserPosition(a.OwnerID, entity.Down, g.initCenterPixelPosition.CenterToTopLeft(user.Character.Size))
		g.trackUserProxy(a.OwnerID, proxyID)
	} else {
		user, err := g.userRepo.GetByID(ctx, a.OwnerID)
		if err != nil {
			return nil, err
		}

		// By default, if user doesn't explicitly choose the character name, we
		// will choose the first one in our list.
		userCharacters, err := g.gameCharacterRepo.GetAllUserCharacters(ctx, user.ID, g.communityID)
		if err != nil {
			return nil, err
		}

		if len(userCharacters) == 0 {
			return nil, errors.New("user must buy a character before")
		}

		var firstCharacter *Character
		for _, uc := range userCharacters {
			if character := g.findCharacterByID(uc.CharacterID); character != nil {
				firstCharacter = character
				break
			}
		}

		if firstCharacter == nil {
			return nil, errors.New("not found any suitable character of user")
		}

		// Create a new user in game state with full information.
		g.addUser(User{
			User: UserInfo{
				ID:        user.ID,
				Name:      user.Name,
				AvatarURL: user.ProfilePicture,
			},
			Character:      firstCharacter,
			Direction:      entity.Down,
			ConnectedBy:    sql.NullString{Valid: true, String: proxyID},
			PixelPosition:  g.initCenterPixelPosition.CenterToTopLeft(firstCharacter.Size),
			LastTimeAction: make(map[string]time.Time),
		})

		for _, uc := range userCharacters {
			character := g.findCharacterByID(uc.CharacterID)
			if character == nil {
				xcontext.Logger(ctx).Errorf("Not found character %s in map", uc.CharacterID)
				continue
			}

			g.trackNewUserCharacter(user.ID, character)
		}
	}

	// Update these fields to serialize to client.
	a.user = *g.userMap[a.OwnerID]
	a.user.PixelPosition = a.user.PixelPosition.CenterToTopLeft(a.user.Character.Size)

	return []model.GameActionServerRequest{{
		UserID: "",
		Type:   InitAction{}.Type(),
		Value:  map[string]any{"to_user_id": a.OwnerID},
	}}, nil
}

////////////////// EXIT Action
type ExitAction struct {
	OwnerID string
}

func (a ExitAction) SendTo() []string {
	// Broadcast to all clients.
	return nil
}

func (a ExitAction) Type() string {
	return "exit"
}

func (a ExitAction) Owner() string {
	return a.OwnerID
}

func (a *ExitAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	user, ok := g.userMap[a.OwnerID]
	if !ok {
		return nil, errors.New("user has not appeared in room")
	}

	if !user.ConnectedBy.Valid {
		return nil, errors.New("the user is inactive, he must not have been appeared in game state")
	}

	g.trackUserProxy(a.OwnerID, "")
	// TODO: This action will reset the position after user exits room.
	// The is using for testing with frontend. If the frontend completed, MUST
	// remove this code.
	g.trackUserPosition(a.OwnerID, entity.Down, g.initCenterPixelPosition.CenterToTopLeft(user.Character.Size))

	return nil, nil
}

////////////////// INIT Action
// InitAction returns to the owner of this action the current game state.
type InitAction struct {
	OwnerID string

	ToUserID       string
	initialUsers   []User
	messageHistory []Message
	luckyboxes     []Luckybox
}

func (a InitAction) SendTo() []string {
	// Send to only the owner of action.
	return []string{a.ToUserID}
}

func (a InitAction) Type() string {
	return "init"
}

func (a InitAction) Owner() string {
	return a.ToUserID
}

func (a *InitAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	if a.OwnerID != "" {
		// Regular user cannot send init action.
		return nil, errors.New("permission denied")
	}

	a.initialUsers = g.SerializeUser()
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
	OwnerID   string
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
	return a.OwnerID
}

func (a *MessageAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	user, ok := g.userMap[a.OwnerID]
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
	OwnerID string
	Emoji   string
}

func (a EmojiAction) SendTo() []string {
	// Send to every one.
	return nil
}

func (a EmojiAction) Type() string {
	return "emoji"
}

func (a EmojiAction) Owner() string {
	return a.OwnerID
}

func (a *EmojiAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	user, ok := g.userMap[a.OwnerID]
	if !ok || !user.ConnectedBy.Valid {
		return nil, errors.New("user is not in map")
	}

	return nil, nil
}

////////////////// START LUCKYBOX EVENT Action
// StartLuckyboxEventAction generates luckybox in room.
type StartLuckyboxEventAction struct {
	OwnerID string
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
	// This action not belongs to any user. Game center triggers it.
	return ""
}

func (a *StartLuckyboxEventAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	if a.OwnerID != "" {
		// Regular user cannot send create_luckybox_event action.
		return nil, errors.New("permission denied")
	}

	event, err := g.gameLuckyboxRepo.GetLuckyboxEventByID(ctx, a.EventID)
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
	OwnerID string
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
	// This action not belongs to any user. Game center triggers it.
	return ""
}

func (a *StopLuckyboxEventAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	if a.OwnerID != "" {
		// Regular user cannot send stop_luckybox_event action.
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
	OwnerID    string
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
	return a.OwnerID
}

func (a *CollectLuckyboxAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	user, ok := g.userMap[a.OwnerID]
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

	err := g.followerRepo.IncreasePoint(ctx, a.OwnerID, g.communityID, uint64(luckybox.Point), false)
	if err != nil {
		return nil, err
	}

	err = g.leaderboard.ChangePointLeaderboard(ctx, int64(luckybox.Point), time.Now(),
		a.OwnerID, g.communityID)
	if err != nil {
		return nil, err
	}

	g.removeLuckybox(luckybox.ID, a.OwnerID)
	a.luckybox = luckybox.WithCenterPixelPosition(g.mapConfig.TileSizeInPixel)

	return nil, nil
}

////////////////// CLEANUP PROXY Action
// CleanupProxyAction is used to clean up disconnected proxy.
type CleanupProxyAction struct {
	OwnerID      string
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
	if a.OwnerID != "" {
		// Regular user cannot send cleanup_proxy action.
		return nil, errors.New("permission denied")
	}

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

////////////////// CHANGE CHARACTER Action
// ChangeCharacterAction changes the current character of user to another one.
type ChangeCharacterAction struct {
	OwnerID     string
	CharacterID string

	userCharacter Character
}

func (a ChangeCharacterAction) SendTo() []string {
	// Send to everyone.
	return nil
}

func (a ChangeCharacterAction) Type() string {
	return "change_character"
}

func (a ChangeCharacterAction) Owner() string {
	return a.OwnerID
}

func (a *ChangeCharacterAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	user, ok := g.userMap[a.OwnerID]
	if !ok {
		return nil, errors.New("user is not in map")
	}

	character := g.findCharacterByID(a.CharacterID)
	if character == nil {
		return nil, fmt.Errorf("not found character %s", a.CharacterID)
	}

	ownedCharacter := user.findOwnedCharacterByID(a.CharacterID)
	if ownedCharacter == nil {
		return nil, fmt.Errorf("user didn't buy the character %s", a.CharacterID)
	}

	g.trackUserCharacter(user.User.ID, character)
	a.userCharacter = *character
	return nil, nil
}

////////////////// CREATE CHARACTER Action
// CreateCharacterAction is used for adding a character to map when super admin
// creates a new one.
type CreateCharacterAction struct {
	OwnerID   string
	Character Character
}

func (a CreateCharacterAction) SendTo() []string {
	// Notify to everyone to update game character at client.
	return nil
}

func (a CreateCharacterAction) Type() string {
	return "create_character"
}

func (a CreateCharacterAction) Owner() string {
	// This action doesn't belong to any owner. Game center triggers it.
	return ""
}

func (a *CreateCharacterAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	if a.OwnerID != "" {
		// Regular user cannot send create_character action.
		return nil, errors.New("permission denied")
	}

	if gameCharacter := g.findCharacterByID(a.Character.ID); gameCharacter != nil {
		// If the character existed, no need to append again.
		*gameCharacter = a.Character
		return nil, nil
	}

	g.characters = append(g.characters, &a.Character)
	return nil, nil
}

////////////////// BUY CHARACTER Action
// BuyCharacterAction is used for adding a character to user when user buys a
// new one.
type BuyCharacterAction struct {
	OwnerID     string
	BuyUserID   string
	CharacterID string
}

func (a BuyCharacterAction) SendTo() []string {
	// Not send this action to anyone.
	return []string{}
}

func (a BuyCharacterAction) Type() string {
	return "buy_character"
}

func (a BuyCharacterAction) Owner() string {
	// This action doesn't belong to any owner. Game center triggers it.
	return ""
}

func (a *BuyCharacterAction) Apply(ctx context.Context, proxyID string, g *GameState) ([]model.GameActionServerRequest, error) {
	if a.OwnerID != "" {
		// Regular user cannot send create_character action.
		return nil, errors.New("permission denied")
	}

	user, ok := g.userMap[a.BuyUserID]
	if !ok {
		return nil, errors.New("user is not in map")
	}

	// If user has already bought this character, no need to track.
	if user.findOwnedCharacterByID(a.CharacterID) != nil {
		return nil, nil
	}

	character := g.findCharacterByID(a.CharacterID)
	if character == nil {
		return nil, fmt.Errorf("not found character %s", a.CharacterID)
	}

	g.trackNewUserCharacter(user.User.ID, character)
	return nil, nil
}
