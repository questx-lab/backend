package gameengine

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
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

	// LastTimeAction is the last time user apply the action. This field is used
	// to prevent user sending actions too fast.
	LastTimeAction map[string]time.Time `json:"-"`

	// IsActive indicates whether user appears on the map.
	IsActive bool `json:"-"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.X, p.Y)
}

func (p Position) distance(another Position) float64 {
	x2 := math.Pow(float64(p.X)-float64(another.X), 2)
	y2 := math.Pow(float64(p.Y)-float64(another.Y), 2)
	return math.Sqrt(x2 + y2)
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

type GamePlayer struct {
	Height int
	Width  int
}

func ParsePlayer(jsonContent []byte) (*GamePlayer, error) {
	m := map[string]any{}
	err := json.Unmarshal(jsonContent, &m)
	if err != nil {
		return nil, err
	}

	frames, ok := m["frames"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid frames")
	}

	player, ok := frames["ariel-back"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid player")
	}

	playerFrame, ok := player["frame"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid player frame")
	}

	w, ok := playerFrame["w"].(float64)
	if !ok {
		return nil, errors.New("invalid width")
	}

	h, ok := playerFrame["h"].(float64)
	if !ok {
		return nil, errors.New("invalid height")
	}

	return &GamePlayer{Height: int(h), Width: int(w)}, nil
}
