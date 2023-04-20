package gameengine

import (
	"fmt"
	"time"

	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/xcontext"
)

const (
	// Player size in pixel.
	// TODO: Need to read this information from tmx map.
	playerWidth  = 32
	playerHeight = 48
)

type GameState struct {
	id     int
	roomID string

	// Width and Height of map in number of tiles (not pixel).
	width  int
	height int

	// Size of a tile (in pixel).
	tileWidth  int
	tileHeight int

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
}

// newGameState creates a game state given a room id.
func newGameState(ctx xcontext.Context, gameRepo repository.GameRepository, roomID string) (*GameState, error) {
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
	parsedMap, err := ParseGameMap(gameMap.Map)
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

	return &GameState{
		roomID:           room.ID,
		width:            parsedMap.Width,
		height:           parsedMap.Height,
		tileWidth:        parsedMap.TileWidth,
		tileHeight:       parsedMap.TileHeight,
		collisionTileMap: collisionTileMap,
		userDiff:         xsync.NewMapOf[*entity.GameUser](),
		gameRepo:         gameRepo,
	}, nil
}

// LoadUser loads all users into game state.
func (g *GameState) LoadUser(ctx xcontext.Context, gameRepo repository.GameRepository) error {
	users, err := gameRepo.GetUsersByRoomID(ctx, g.roomID)
	if err != nil {
		return err
	}

	g.userMap = make(map[string]*User)
	for _, user := range users {
		userPixelPosition := Position{X: user.PositionX, Y: user.PositionY}
		if g.isObjectCollision(userPixelPosition, playerWidth, playerHeight) {
			return fmt.Errorf("detected a user standing on a collision tile at pixel %s", userPixelPosition)
		}

		g.addUser(User{
			UserID:        user.UserID,
			Direction:     user.Direction,
			PixelPosition: userPixelPosition,
			LastTimeMoved: time.Now(),
			IsActive:      user.IsActive,
		})
	}

	return nil
}

// Apply applies an action into game state. The game state will save this action
// as history to revert if needed. This method returns the new game state id.
func (g *GameState) Apply(action Action) (int, error) {
	if err := action.Apply(g); err != nil {
		return 0, err
	}

	// Actions sent to only owner should not change the game state, so the id of
	// game state won't be changed.
	if !action.OnlyOwner() {
		g.id++
	}

	return g.id, nil
}

// Serialize returns a bytes object in JSON format representing for current
// position of all users.
func (g *GameState) Serialize() SerializedGameState {
	var users []User
	for _, user := range g.userMap {
		if user.IsActive {
			users = append(users, *user)
		}
	}

	return SerializedGameState{
		ID:    g.id,
		Users: users,
	}
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
func (g *GameState) trackUserPosition(userID string, position Position) {
	diff := g.loadOrStoreDiff(userID)
	if diff == nil {
		return
	}

	diff.PositionX = position.X
	diff.PositionY = position.Y

	g.userMap[userID].PixelPosition = position
	g.userMap[userID].LastTimeMoved = time.Now()
}

// updateUserPositionDiff tracks the direction of user to update in database.
func (g *GameState) trackUserDirection(userID string, direction entity.DirectionType) {
	diff := g.loadOrStoreDiff(userID)
	if diff == nil {
		return
	}

	diff.Direction = direction
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
	tilePosition := g.pixelToTile(pointPixel)
	_, isBlocked := g.collisionTileMap[tilePosition]
	if isBlocked {
		return true
	}

	if tilePosition.X < 0 || tilePosition.X >= g.width {
		return true
	}

	if tilePosition.Y < 0 || tilePosition.Y >= g.height {
		return true
	}

	return false
}

// pixelToTile returns position in tile given a position in pixel.
func (g *GameState) pixelToTile(p Position) Position {
	return Position{X: p.X / g.tileWidth, Y: p.Y / g.tileHeight}
}
