package gameengine

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/entity"
)

type User struct {
	UserID string `json:"user_id"`

	// If the user presses the moving button which is the same with user's
	// direction, the game state treats it as a moving action.
	//
	// If the user presses a moving button which is difference from user's
	// direction, the game state treats it as rotating action.
	Direction entity.DirectionType `json:"direction"`

	// PixelPosition is the position in pixel of user.
	PixelPosition Position `json:"pixel_position"`

	// LastTimeMoved is the last time user uses the Moving Action. This is used
	// to track the moving speed of user.
	//
	// For example, if the last time moved of user is 10h12m9s, then the next
	// time user can move is 10h12m10s.
	LastTimeMoved time.Time `json:"last_time_moved"`

	// IsActive indicates whether user appears on the map.
	IsActive bool `json:"is_active"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.X, p.Y)
}

func (p Position) move(direction entity.DirectionType) Position {
	switch direction {
	case entity.Up:
		return Position{X: p.X, Y: p.Y - 1}
	case entity.Down:
		return Position{X: p.X, Y: p.Y + 1}
	case entity.Right:
		return Position{X: p.X + 1, Y: p.Y}
	case entity.Left:
		return Position{X: p.X - 1, Y: p.Y}
	}

	return p
}

const collisionValue = float64(40)

type GameMap struct {
	Height         int
	Width          int
	TileHeight     int
	TileWidth      int
	CollisionLayer [][]bool
}

func ParseGameMap(jsonContent []byte) (*GameMap, error) {
	m := map[string]any{}
	err := json.Unmarshal(jsonContent, &m)
	if err != nil {
		return nil, err
	}

	height, ok := m["height"].(float64)
	if !ok {
		return nil, errors.New("invalid map height")
	}

	width, ok := m["width"].(float64)
	if !ok {
		return nil, errors.New("invalid map width")
	}

	tileHeight, ok := m["tileheight"].(float64)
	if !ok {
		return nil, errors.New("invalid map tileHeight")
	}

	tileWidth, ok := m["tilewidth"].(float64)
	if !ok {
		return nil, errors.New("invalid map tileWidth")
	}

	gameMap := GameMap{
		Height:     int(height),
		Width:      int(width),
		TileHeight: int(tileHeight),
		TileWidth:  int(tileWidth),
	}

	layers, ok := m["layers"].([]any)
	if !ok {
		return nil, errors.New("invalid map layers")
	}

	for _, layer := range layers {
		mapLayer, ok := layer.(map[string]any)
		if !ok {
			return nil, errors.New("invalid map layer")
		}

		if name, ok := mapLayer["name"]; ok && name == "CollisionLayer" {
			data, ok := mapLayer["data"].([]any)
			if !ok {
				return nil, errors.New("invalid collision layer data")
			}

			if len(data) != gameMap.Width*gameMap.Height {
				return nil, errors.New("invalid number of elements in collision layer data")
			}

			gameMap.CollisionLayer = make([][]bool, gameMap.Width)
			for i := range gameMap.CollisionLayer {
				gameMap.CollisionLayer[i] = make([]bool, gameMap.Height)
			}

			for i := range data {
				gameMap.CollisionLayer[i%gameMap.Width][i/gameMap.Width] = data[i] == collisionValue
			}
		}
	}

	if len(gameMap.CollisionLayer) == 0 {
		return nil, errors.New("not found collision layer")
	}

	return &gameMap, nil
}

type SerializedGameState struct {
	ID    int
	Users []User
}
