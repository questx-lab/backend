package gameengine

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
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
	playerWidth  int
	playerHeight int

	// Initial position if user hadn't joined the room yet.
	initialPosition Position

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

	// actionDelay indicates how long the action can be applied again.
	actionDelay map[string]time.Duration

	// messageHistory stores last messages of game.
	messageHistory []Message
}

// newGameState creates a game state given a room id.
func newGameState(ctx context.Context, gameRepo repository.GameRepository, roomID string) (*GameState, error) {
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

	// Parse tmx map content from game map.
	parsedMap, err := ParseGameMap(gameMap.Map, strings.Split(gameMap.CollisionLayers, ","))
	if err != nil {
		return nil, err
	}

	parsedPlayer, err := ParsePlayer(gameMap.Player)
	if err != nil {
		return nil, err
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
		playerWidth:      parsedPlayer.Width,
		playerHeight:     parsedPlayer.Height,
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

	centerPosition := Position{gameMap.InitX, gameMap.InitY}
	gamestate.initialPosition = centerPosition.centerToTopLeft(gamestate.playerWidth, gamestate.playerHeight)
	if gamestate.isObjectCollision(gamestate.initialPosition, gamestate.playerWidth, gamestate.playerHeight) {
		return nil, errors.New("initial game state is standing on a collision object")
	}

	return gamestate, nil
}

// LoadUser loads all users into game state.
func (g *GameState) LoadUser(ctx context.Context, gameRepo repository.GameRepository) error {
	users, err := gameRepo.GetUsersByRoomID(ctx, g.roomID)
	if err != nil {
		return err
	}

	g.userMap = make(map[string]*User)
	for _, user := range users {
		userPixelPosition := Position{X: user.PositionX, Y: user.PositionY}
		if g.isObjectCollision(userPixelPosition, g.playerWidth, g.playerHeight) {
			xcontext.Logger(ctx).Errorf("Detected a user standing on a collision tile at pixel %s", userPixelPosition)
			continue
		}

		g.addUser(User{
			UserID:         user.UserID,
			Direction:      user.Direction,
			PixelPosition:  userPixelPosition,
			LastTimeAction: make(map[string]time.Time),
			IsActive:       user.IsActive,
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
			clientUser.PixelPosition = clientUser.PixelPosition.topLeftToCenter(g.playerWidth, g.playerHeight)
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

	gameUser, _ := g.userDiff.LoadOrStore(user.UserID, &entity.GameUser{
		UserID:    user.UserID,
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
	g.userDiff.Store(user.UserID, &entity.GameUser{
		UserID:    user.UserID,
		RoomID:    g.roomID,
		PositionX: user.PixelPosition.X,
		PositionY: user.PixelPosition.Y,
		Direction: user.Direction,
		IsActive:  user.IsActive,
	})

	g.userMap[user.UserID] = &user
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
