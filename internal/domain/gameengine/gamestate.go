package gameengine

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/domain/statistic"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type GameState struct {
	roomID      string
	communityID string

	// Width and Height of map in number of tiles (not pixel).
	mapConfig *GameMap

	// Size of player (in pixel).
	players []Player

	// Initial position if user hadn't joined the room yet.
	initCenterPos Position

	// userDiff contains all user differences between the original game state vs
	// the current game state.
	// DO NOT modify this field directly, please use setter methods instead.
	userDiff *xsync.MapOf[string, *entity.GameUser]

	// userMap contains user information in this game. It uses pixel unit to
	// determine its position.
	userMap map[string]*User

	gameRepo     repository.GameRepository
	userRepo     repository.UserRepository
	followerRepo repository.FollowerRepository
	leaderboard  statistic.Leaderboard

	// actionDelay indicates how long the action can be applied again.
	actionDelay map[string]time.Duration

	// messageHistory stores last messages of game.
	messageHistory []Message

	// luckybox information.
	luckyboxes           map[string]Luckybox
	luckyboxesByPosition map[Position]Luckybox

	// luckyboxDiff contains all luckybox differences between the original game
	// state vs the current game state.
	// DO NOT modify this field directly, please use setter methods instead.
	luckyboxDiff *xsync.MapOf[string, *entity.GameLuckybox]
}

// newGameState creates a game state given a room id.
func newGameState(
	ctx context.Context,
	gameRepo repository.GameRepository,
	userRepo repository.UserRepository,
	followerRepo repository.FollowerRepository,
	leaderboard statistic.Leaderboard,
	storage storage.Storage,
	roomID string,
) (*GameState, error) {
	// Get room information from room id.
	room, err := gameRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// Get map information from map id.
	gameMap, err := gameRepo.GetMapByID(ctx, room.MapID)
	if err != nil {
		return nil, err
	}

	mapConfig, err := storage.DownloadFromURL(ctx, gameMap.ConfigURL)
	if err != nil {
		return nil, err
	}

	// Parse tmx map content from game map.
	parsedMap, err := ParseGameMap(mapConfig, strings.Split(gameMap.CollisionLayers, ","))
	if err != nil {
		return nil, err
	}

	players, err := gameRepo.GetPlayersByMapID(ctx, gameMap.ID)
	if err != nil {
		return nil, err
	}

	var playerList []Player
	for _, player := range players {
		playerConfig, err := storage.DownloadFromURL(ctx, player.ConfigURL)
		if err != nil {
			return nil, err
		}

		parsedPlayer, err := ParsePlayer(playerConfig)
		if err != nil {
			return nil, err
		}

		playerList = append(playerList, Player{
			ID:     player.ID,
			Name:   player.Name,
			Width:  parsedPlayer.Width,
			Height: parsedPlayer.Height,
		})
	}

	gameCfg := xcontext.Configs(ctx).Game
	gamestate := &GameState{
		roomID:         room.ID,
		communityID:    room.CommunityID,
		mapConfig:      parsedMap,
		players:        playerList,
		userDiff:       xsync.NewMapOf[*entity.GameUser](),
		gameRepo:       gameRepo,
		userRepo:       userRepo,
		followerRepo:   followerRepo,
		leaderboard:    leaderboard,
		messageHistory: make([]Message, 0, gameCfg.MessageHistoryLength),
		actionDelay: map[string]time.Duration{
			MoveAction{}.Type():            gameCfg.MoveActionDelay,
			InitAction{}.Type():            gameCfg.InitActionDelay,
			JoinAction{}.Type():            gameCfg.JoinActionDelay,
			MessageAction{}.Type():         gameCfg.MessageActionDelay,
			CollectLuckyboxAction{}.Type(): gameCfg.CollectLuckyboxActionDelay,
		},
	}

	for _, player := range playerList {
		gamestate.initCenterPos = Position{gameMap.InitX, gameMap.InitY}
		topLeftInitPos := gamestate.initCenterPos.CenterToTopLeft(player)
		if gamestate.mapConfig.IsPlayerCollision(topLeftInitPos, player) {
			return nil, fmt.Errorf("initial of player %s is standing on a collision object", player.Name)
		}
	}

	return gamestate, nil
}

// LoadUser loads all users into game state.
func (g *GameState) LoadUser(ctx context.Context) error {
	users, err := g.gameRepo.GetUsersByRoomID(ctx, g.roomID)
	if err != nil {
		return err
	}

	g.userMap = make(map[string]*User)
	for _, gameUser := range users {
		player := g.findPlayerByID(gameUser.GamePlayerID)
		userPixelPosition := Position{X: gameUser.PositionX, Y: gameUser.PositionY}
		if g.mapConfig.IsPlayerCollision(userPixelPosition, player) {
			xcontext.Logger(ctx).Errorf("Detected a user standing on a collision tile at pixel %s", userPixelPosition)
			continue
		}

		user, err := g.userRepo.GetByID(ctx, gameUser.UserID)
		if err != nil {
			return err
		}

		g.addUser(User{
			User: UserInfo{
				ID:        user.ID,
				Name:      user.Name,
				AvatarURL: user.ProfilePicture,
			},
			Player:         player,
			Direction:      gameUser.Direction,
			PixelPosition:  userPixelPosition,
			LastTimeAction: make(map[string]time.Time),
			IsActive:       gameUser.IsActive,
		})
	}

	return nil
}

// LoadUser loads all users into game state.
func (g *GameState) LoadLuckybox(ctx context.Context) error {
	luckyboxes, err := g.gameRepo.GetAvailableLuckyboxesByRoomID(ctx, g.roomID)
	if err != nil {
		return err
	}

	g.luckyboxes = make(map[string]Luckybox)
	for _, luckybox := range luckyboxes {
		luckyboxState := Luckybox{
			ID:      luckybox.ID,
			EventID: luckybox.EventID,
			Point:   luckybox.Point,
			Position: Position{
				X: luckybox.PositionX,
				Y: luckybox.PositionY,
			},
		}

		if _, ok := g.mapConfig.CollisionTileMap[luckyboxState.Position]; ok {
			xcontext.Logger(ctx).Errorf("Luckybox %s appears on collision layer", luckyboxState.ID)
			continue
		}

		if another, ok := g.luckyboxesByPosition[luckyboxState.Position]; ok {
			xcontext.Logger(ctx).Errorf("Luckybox %s overlaps on %s", luckyboxState.ID, another.ID)
			continue
		}

		g.addLuckybox(luckyboxState)
	}

	return nil
}

// Apply applies an action into game state.
func (g *GameState) Apply(ctx context.Context, action Action) error {
	if delay, ok := g.actionDelay[action.Type()]; ok {
		if user, ok := g.userMap[action.Owner()]; ok {
			if last, ok := user.LastTimeAction[action.Type()]; ok && time.Since(last) < delay {
				return fmt.Errorf("submit action %s too fast", action.Type())
			}
		}
	}

	if err := action.Apply(ctx, g); err != nil {
		return err
	}

	if user, ok := g.userMap[action.Owner()]; ok {
		user.LastTimeAction[action.Type()] = time.Now()
	}

	return nil
}

// Serialize returns a bytes object in JSON format representing for current
// position of all users.
func (g *GameState) Serialize() []User {
	var users []User
	for _, user := range g.userMap {
		if user.IsActive {
			clientUser := *user
			clientUser.PixelPosition = clientUser.PixelPosition.TopLeftToCenter(user.Player)
			users = append(users, clientUser)
		}
	}

	return users
}

// UserDiff returns all database tracking differences of game user until now.
// The diff will be reset after this method is called.
//
// Usage example:
//
//   for _, user := range gamestate.UserDiff() {
//       gameRepo.UpdateGameUser(ctx, user)
//   }
func (g *GameState) UserDiff() []*entity.GameUser {
	diff := []*entity.GameUser{}
	g.userDiff.Range(func(key string, value *entity.GameUser) bool {
		diff = append(diff, value)
		g.userDiff.Delete(key)
		return true
	})

	return diff
}

// LuckyboxDiff returns all database tracking differences of game luckybox until
// now. The diff will be reset after this method is called.
//
// Usage example:
//
//   for _, luckybox := range gamestate.LuckyboxDiff() {
//       gameRepo.UpdateLuckybox(ctx, luckybox)
//   }
func (g *GameState) LuckyboxDiff() []*entity.GameLuckybox {
	diff := []*entity.GameLuckybox{}
	g.luckyboxDiff.Range(func(key string, value *entity.GameLuckybox) bool {
		diff = append(diff, value)
		g.luckyboxDiff.Delete(key)
		return true
	})

	return diff
}

// trackUserPosition tracks the position of user to update in database.
func (g *GameState) trackUserPosition(userID string, direction entity.DirectionType, position Position) {
	diff := g.loadOrStoreUserDiff(userID)
	if diff == nil {
		return
	}

	diff.PositionX = position.X
	diff.PositionY = position.Y
	diff.Direction = direction

	g.userMap[userID].PixelPosition = position
	g.userMap[userID].Direction = direction
}

// trackUserActive tracks the status of user to update in database.
func (g *GameState) trackUserActive(userID string, isActive bool) {
	diff := g.loadOrStoreUserDiff(userID)
	if diff == nil {
		return
	}

	diff.IsActive = isActive
	g.userMap[userID].IsActive = isActive
}

func (g *GameState) loadOrStoreUserDiff(userID string) *entity.GameUser {
	user, ok := g.userMap[userID]
	if !ok {
		return nil
	}

	gameUser, _ := g.userDiff.LoadOrStore(user.User.ID, &entity.GameUser{
		UserID:       user.User.ID,
		RoomID:       g.roomID,
		GamePlayerID: user.Player.ID,
		PositionX:    user.PixelPosition.X,
		PositionY:    user.PixelPosition.Y,
		Direction:    user.Direction,
		IsActive:     user.IsActive,
	})

	return gameUser
}

// addUser creates a new user in room.
func (g *GameState) addUser(user User) {
	g.userDiff.Store(user.User.ID, &entity.GameUser{
		UserID:       user.User.ID,
		RoomID:       g.roomID,
		GamePlayerID: user.Player.ID,
		PositionX:    user.PixelPosition.X,
		PositionY:    user.PixelPosition.Y,
		Direction:    user.Direction,
		IsActive:     user.IsActive,
	})

	g.userMap[user.User.ID] = &user
}

// removeLuckybox marks the luckybox as collected.
func (g *GameState) removeLuckybox(luckyboxID string, userID string) {
	luckybox, ok := g.luckyboxes[luckyboxID]
	if !ok {
		return
	}

	delete(g.luckyboxes, luckyboxID)
	delete(g.luckyboxesByPosition, luckybox.Position)

	collectedBy := sql.NullString{Valid: false}
	if userID != "" {
		collectedBy = sql.NullString{Valid: true, String: userID}
	}

	g.luckyboxDiff.Store(luckybox.ID, &entity.GameLuckybox{
		Base:        entity.Base{ID: luckybox.ID},
		EventID:     luckybox.EventID,
		PositionX:   luckybox.Position.X,
		PositionY:   luckybox.Position.Y,
		Point:       luckybox.Point,
		CollectedBy: collectedBy,
	})
}

// addLuckybox creates a new luckybox in room.
func (g *GameState) addLuckybox(luckybox Luckybox) {
	g.luckyboxDiff.Store(luckybox.ID, &entity.GameLuckybox{
		Base:        entity.Base{ID: luckybox.ID},
		EventID:     luckybox.EventID,
		PositionX:   luckybox.Position.X,
		PositionY:   luckybox.Position.Y,
		Point:       luckybox.Point,
		CollectedBy: sql.NullString{},
	})

	g.luckyboxes[luckybox.ID] = luckybox
	g.luckyboxesByPosition[luckybox.Position] = luckybox
}

func (g *GameState) findPlayerByID(id string) Player {
	for _, p := range g.players {
		if p.ID == id {
			return p
		}
	}

	return g.players[0]
}
