package gamestate

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/azul3d/engine/tmx"
)

const (
	maxActionHistoryLength = 10

	// The value represents for a collision tile.
	collisionValue = 40

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

	actionHistory []Action

	// Diff contains all differences between the original game state vs the
	// current game state.
	// DO NOT modify this field directly, please use setter methods instead.
	diff map[string]any

	// collisionTileMap indicates which tile is collision.
	collisionTileMap map[Position]any

	// userMap contains user information in this game. It uses pixel unit to
	// determine its position.
	userMap map[string]User

	gameRepo repository.GameRepository
}

// New creates a game state given a room id.
func New(ctx xcontext.Context, gameRepo repository.GameRepository, roomID string) (*GameState, error) {
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
	tmxMap, err := tmx.Parse(gameMap.Content)
	if err != nil {
		return nil, err
	}

	// Find the collision layer to extract collision tiles.
	var collisionLayer *tmx.Layer
	for _, layer := range tmxMap.Layers {
		if layer.Name == "CollisionLayer" {
			collisionLayer = layer
			break
		}
	}

	if collisionLayer == nil {
		return nil, errors.New("not found collision layer")
	}

	collisionTileMap := make(map[Position]any)
	for coord, value := range collisionLayer.Tiles {
		if value == collisionValue {
			collisionTileMap[Position{X: coord.X, Y: coord.Y}] = nil
		}
	}

	return &GameState{
		roomID:           room.ID,
		width:            tmxMap.Width,
		height:           tmxMap.Height,
		tileWidth:        tmxMap.TileWidth,
		tileHeight:       tmxMap.TileHeight,
		collisionTileMap: collisionTileMap,
		diff:             make(map[string]any),
		gameRepo:         gameRepo,
	}, nil
}

// LoadUser loads all users into game state.
func (g *GameState) LoadUser(ctx xcontext.Context, gameRepo repository.GameRepository) error {
	users, err := gameRepo.GetUsersByRoomID(ctx, g.roomID)
	if err != nil {
		return err
	}

	g.userMap = make(map[string]User)
	for _, user := range users {
		if !user.IsActive {
			continue
		}

		userPixelPosition := Position{X: user.PositionX, Y: user.PositionY}
		if g.isObjectCollision(userPixelPosition, playerWidth, playerHeight) {
			return fmt.Errorf("detected a user standing on a collision tile at pixel %s", userPixelPosition)
		}

		g.userMap[user.UserID] = User{
			UserID:        user.UserID,
			Direction:     user.Direction,
			PixelPosition: userPixelPosition,
			LastTimeMoved: time.Now(),
		}
	}

	return nil
}

// Apply applies an action into game state. The game state will save this action
// as history to revert if needed. This method returns the new game state id.
func (g *GameState) Apply(action Action) (int, error) {
	if err := action.Apply(g); err != nil {
		return 0, err
	}

	g.id++
	g.actionHistory = append([]Action{action}, g.actionHistory...)
	if len(g.actionHistory) >= maxActionHistoryLength {
		g.actionHistory = g.actionHistory[:maxActionHistoryLength]
	}

	return g.id, nil
}

func (g *GameState) Summon(userID string) (User, error) {
	user, ok := g.userMap[userID]
	if ok {
		return user, nil
	}

	g.userMap[userID] = User{
		UserID:        userID,
		Direction:     entity.Down,
		PixelPosition: Position{0, 0},
		LastTimeMoved: time.Now(),
	}

	return g.userMap[userID], nil
}

// Serialize returns a bytes object in JSON format representing for current
// position of all users.
func (g *GameState) Serialize() ([]byte, error) {
	data := map[string]any{
		"id":    g.id,
		"users": g.userMap,
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Diff returns all database tracking differences until now. The diff will be
// reset after this method is called.
//
// Usage example:
//
//   for _, obj := range gamestate.Diff() {
//       switch t := obj.(type) {
//	     case entity.GameUser:
//          gameRepo.UpdateGameUser(ctx, t)
//       }
//   }
func (g *GameState) Diff() []any {
	diff := []any{}
	for _, v := range g.diff {
		diff = append(diff, v)
	}

	g.diff = make(map[string]any)
	return diff
}

// updateUserDiff should be called by action.Apply() method when the user is
// updated to another position or direction. This method is used to keep track
// database difference.
func (g *GameState) updateUserDiff(user entity.GameUser) {
	g.diff["user_"+user.UserID] = user
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
	_, isBlocked := g.collisionTileMap[g.pixelToTile(pointPixel)]
	return isBlocked
}

// pixelToTile returns position in tile given a position in pixel.
func (g *GameState) pixelToTile(p Position) Position {
	return Position{X: p.X / g.tileWidth, Y: p.Y / g.tileHeight}
}
