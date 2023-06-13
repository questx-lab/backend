package gameengine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type GameState struct {
	roomID string

	// Width and Height of map in number of tiles (not pixel).
	width  int
	height int

	// Size of a tile (in pixel).
	tileWidth  int
	tileHeight int

	// Size of player (in pixel).
	players []Player

	// Initial position if user hadn't joined the room yet.
	initCentrPos Position

	// userDiff contains all user differences between the original game state vs
	// the current game state.
	// DO NOT modify this field directly, please use setter methods instead.
	userDiff *xsync.MapOf[string, *entity.GameUser]

	// collisionTileMap indicates which tile is collision.
	collisionTileMap map[Position]any

	// userMap contains user information in this game. It uses pixel unit to
	// determine its position.
	userMap map[string]*User

	gameRepo repository.GameRepository
	userRepo repository.UserRepository

	// actionDelay indicates how long the action can be applied again.
	actionDelay map[string]time.Duration

	// messageHistory stores last messages of game.
	messageHistory []Message
}

// newGameState creates a game state given a room id.
func newGameState(
	ctx context.Context,
	gameRepo repository.GameRepository,
	userRepo repository.UserRepository,
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

	players, err := gameRepo.GetPlayerByMapID(ctx, gameMap.ID)
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

	collisionTileMap := make(map[Position]any)
	for i := range parsedMap.CollisionLayer {
		for j := range parsedMap.CollisionLayer[i] {
			if parsedMap.CollisionLayer[i][j] {
				collisionTileMap[Position{X: i, Y: j}] = nil
			}
		}
	}

	gameCfg := xcontext.Configs(ctx).Game
	gamestate := &GameState{
		roomID:           room.ID,
		width:            parsedMap.Width,
		height:           parsedMap.Height,
		tileWidth:        parsedMap.TileWidth,
		tileHeight:       parsedMap.TileHeight,
		players:          playerList,
		collisionTileMap: collisionTileMap,
		userDiff:         xsync.NewMapOf[*entity.GameUser](),
		gameRepo:         gameRepo,
		messageHistory:   make([]Message, 0, gameCfg.MessageHistoryLength),
		actionDelay: map[string]time.Duration{
			MoveAction{}.Type(): gameCfg.MoveActionDelay,
			InitAction{}.Type(): gameCfg.InitActionDelay,
			JoinAction{}.Type(): gameCfg.JoinActionDelay,
		},
	}

	for _, player := range playerList {
		gamestate.initCentrPos = Position{gameMap.InitX, gameMap.InitY}
		topLeftInitPos := gamestate.initCentrPos.centerToTopLeft(player)
		if gamestate.isObjectCollision(topLeftInitPos, player.Width, player.Height) {
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
		if g.isObjectCollision(userPixelPosition, player.Width, player.Height) {
			xcontext.Logger(ctx).Errorf("Detected a user standing on a collision tile at pixel %s", userPixelPosition)
			continue
		}

		user, err := g.userRepo.GetByID(ctx, gameUser.UserID)
		if err != nil {
			return err
		}

		g.addUser(User{
			User: model.User{
				ID:           user.ID,
				Name:         user.Name,
				AvatarURL:    user.ProfilePicture,
				ReferralCode: user.ReferralCode,
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
			clientUser.PixelPosition = clientUser.PixelPosition.topLeftToCenter(user.Player)
			users = append(users, *user)
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

// trackUserPosition tracks the position of user to update in database.
func (g *GameState) trackUserPosition(userID string, direction entity.DirectionType, position Position) {
	diff := g.loadOrStoreDiff(userID)
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
	diff := g.loadOrStoreDiff(userID)
	if diff == nil {
		return
	}

	diff.IsActive = isActive
	g.userMap[userID].IsActive = isActive
}

func (g *GameState) loadOrStoreDiff(userID string) *entity.GameUser {
	user, ok := g.userMap[userID]
	if !ok {
		return nil
	}

	gameUser, _ := g.userDiff.LoadOrStore(user.User.ID, &entity.GameUser{
		UserID:    user.User.ID,
		RoomID:    g.roomID,
		PositionX: user.PixelPosition.X,
		PositionY: user.PixelPosition.Y,
		Direction: user.Direction,
		IsActive:  user.IsActive,
	})

	return gameUser
}

// addUser creates a new user in room.
func (g *GameState) addUser(user User) {
	g.userDiff.Store(user.User.ID, &entity.GameUser{
		UserID:    user.User.ID,
		RoomID:    g.roomID,
		PositionX: user.PixelPosition.X,
		PositionY: user.PixelPosition.Y,
		Direction: user.Direction,
		IsActive:  user.IsActive,
	})

	g.userMap[user.User.ID] = &user
}

// isObjectCollision checks if the object is collided with any collision tile or
// not. The object is represented by its center point, width, and height. All
// parameters must be in pixel.
func (g *GameState) isObjectCollision(topLeftInPixel Position, widthPixel, heightPixel int) bool {
	if g.isPointCollision(topLeftInPixel) {
		return true
	}

	if g.isPointCollision(topRight(topLeftInPixel, widthPixel, heightPixel)) {
		return true
	}

	if g.isPointCollision(bottomLeft(topLeftInPixel, widthPixel, heightPixel)) {
		return true
	}

	if g.isPointCollision(bottomRight(topLeftInPixel, widthPixel, heightPixel)) {
		return true
	}

	return false
}

// isPointCollision checks if a point is collided with any collision tile or
// not. The point position must be in pixel.
func (g *GameState) isPointCollision(pointPixel Position) bool {
	if pointPixel.X < 0 || pointPixel.Y < 0 {
		return true
	}

	tilePosition := g.pixelToTile(pointPixel)
	_, isBlocked := g.collisionTileMap[tilePosition]
	if isBlocked {
		return true
	}

	if tilePosition.X >= g.width || tilePosition.Y >= g.height {
		return true
	}

	return false
}

// pixelToTile returns position in tile given a position in pixel.
func (g *GameState) pixelToTile(p Position) Position {
	return Position{X: p.X / g.tileWidth, Y: p.Y / g.tileHeight}
}

func (g *GameState) findPlayerByName(name string) Player {
	for _, p := range g.players {
		if p.Name == name {
			return p
		}
	}

	return g.players[0]
}

func (g *GameState) findPlayerByID(id string) Player {
	for _, p := range g.players {
		if p.ID == id {
			return p
		}
	}

	return g.players[0]
}
