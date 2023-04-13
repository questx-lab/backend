package gamestate

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/xcontext"
)

const (
	MaxActionHistoryLength = 10

	// CellSize is the size in pixel of a cell.
	CellSize = 4

	// Player size.
	PlayerWidth  = 4
	PlayerHeight = 6
)

type GameState struct {
	id int

	width  int
	height int

	actionHistory []Action

	// Diff contains all differences between the original game state vs the
	// current game state.
	// DO NOT modify this field directly, please use setter methods instead.
	diff map[string]any

	// blockedCellMap indicates which cell is blocked. It uses cell unit to
	// determine its position.
	blockedCellMap map[Position]any

	// userMap uses pixel unit to determine its position.
	userMap        map[Position]User
	userMapReverse map[string]Position
}

// New creates a game state given a room id.
func New(ctx xcontext.Context, gameRepo repository.GameRepository, roomID string) (*GameState, error) {
	room, err := gameRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	gameMap, err := gameRepo.GetMapByID(ctx, room.MapID)
	if err != nil {
		return nil, err
	}

	blockedCells, err := gameRepo.GetBlockedCellsByMapID(ctx, room.MapID)
	if err != nil {
		return nil, err
	}

	users, err := gameRepo.GetUsersByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	cellMap := make(map[Position]any)
	userMap := make(map[Position]User)
	userReverseMap := make(map[string]Position)

	for _, cell := range blockedCells {
		position := Position{X: cell.PositionX, Y: cell.PositionY}
		if _, ok := cellMap[position]; ok {
			return nil, fmt.Errorf("detected overlapping blocked cell at %s", position)
		}

		cellMap[position] = nil
	}

	for _, user := range users {
		if !user.IsActive {
			continue
		}

		userPosition := Position{X: user.PositionX, Y: user.PositionY}
		if _, ok := userMap[userPosition]; ok {
			return nil, fmt.Errorf("detected overlapping users at %s", userPosition)
		}

		cellPosition := userPosition.pixelToCell()
		if _, ok := cellMap[cellPosition]; ok {
			return nil, fmt.Errorf("detected a user standing on a blocked cell at %s", cellPosition)
		}

		userMap[userPosition] = User{
			UserID:        user.UserID,
			Direction:     user.Direction,
			LastTimeMoved: time.Now(),
		}

		userReverseMap[user.UserID] = userPosition
	}

	return &GameState{
		width:          gameMap.Width,
		height:         gameMap.Height,
		blockedCellMap: cellMap,
		userMap:        userMap,
		userMapReverse: userReverseMap,
	}, nil
}

// Clone creates a new game state with deep copies of inner objects.
func (g *GameState) Clone() *GameState {
	clone := *g
	clone.blockedCellMap = make(map[Position]any)
	clone.userMap = make(map[Position]User)
	clone.userMapReverse = make(map[string]Position)
	clone.diff = make(map[string]any)

	for k := range g.blockedCellMap {
		clone.blockedCellMap[k] = nil
	}

	for k, v := range g.userMap {
		clone.userMap[k] = v
	}

	for k, v := range g.userMapReverse {
		clone.userMapReverse[k] = v
	}

	for k, v := range g.diff {
		clone.diff[k] = v
	}

	return &clone
}

// Apply applies an action into game state. The game state will save this action
// as history to revert if needed. This method returns the new game state id.
func (g *GameState) Apply(action Action) (int, error) {
	if err := action.Apply(g); err != nil {
		return 0, err
	}

	g.id++
	g.actionHistory = append([]Action{action}, g.actionHistory...)
	if len(g.actionHistory) >= MaxActionHistoryLength {
		g.actionHistory = g.actionHistory[:MaxActionHistoryLength]
	}

	return g.id, nil
}

// RevertTo returns a reverted the game state to a specified id. The reverted
// state only affects on Serialize() method. Please use other methods with a
// CAUTION.
func (g *GameState) RevertTo(id int) (*GameState, error) {
	if g.id < id {
		return nil, errors.New("unreached game state id")
	}

	if g.id-id > MaxActionHistoryLength {
		return nil, errors.New("too old game state id")
	}

	clone := g.Clone()
	for _, action := range g.actionHistory[:g.id-id] {
		if err := action.Revert(clone); err != nil {
			return nil, err
		}
	}

	return clone, nil
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

// isObjectCollision checks if the object is collided with any blocked cell or
// not. The object is represented by its center point, width, and height. All
// parameters must be in pixel.
func (g *GameState) isObjectCollision(center Position, width, height int) bool {
	if g.isPointCollision(topLeft(center, width, height)) {
		return true
	}

	if g.isPointCollision(topRight(center, width, height)) {
		return true
	}

	if g.isPointCollision(bottomLeft(center, width, height)) {
		return true
	}

	if g.isPointCollision(bottomRight(center, width, height)) {
		return true
	}

	return true
}

// isPointCollision checks if a point is collided with any blocked cell or not.
// The point position must be in pixel.
func (g *GameState) isPointCollision(point Position) bool {
	_, isBlocked := g.blockedCellMap[point.pixelToCell()]
	return isBlocked
}
