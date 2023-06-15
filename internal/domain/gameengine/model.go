package gameengine

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"golang.org/x/exp/slices"
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

func (p Position) centerToTopLeft(width, height int) Position {
	return Position{p.X - width/2, p.Y - height/2}
}

func (p Position) topLeftToCenter(width, height int) Position {
	return Position{p.X + width/2, p.Y + height/2}
}

type GameMap struct {
	Height         int
	Width          int
	TileHeight     int
	TileWidth      int
	CollisionLayer [][]bool
}

func ParseGameMap(jsonContent []byte, collisionLayers []string) (*GameMap, error) {
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

	gameMap.CollisionLayer = make([][]bool, gameMap.Width)
	for i := range gameMap.CollisionLayer {
		gameMap.CollisionLayer[i] = make([]bool, gameMap.Height)
	}

	for _, layer := range layers {
		mapLayer, ok := layer.(map[string]any)
		if !ok {
			return nil, errors.New("invalid map layer")
		}

		if name, ok := mapLayer["name"].(string); ok && slices.Contains(collisionLayers, name) {
			pos := slices.Index(collisionLayers, name)
			collisionLayers = slices.Delete(collisionLayers, pos, pos+1)

			layerType, ok := mapLayer["type"].(string)
			if !ok {
				return nil, fmt.Errorf("not found layer type of %s", name)
			}

			if layerType == "tilelayer" {
				data, ok := mapLayer["data"].([]any)
				if !ok {
					return nil, fmt.Errorf("invalid collision layer data %s", name)
				}

				if len(data) != gameMap.Width*gameMap.Height {
					return nil, fmt.Errorf("invalid number of elements in collision layer data %s", name)
				}

				for i := range data {
					x := i % gameMap.Width
					y := i / gameMap.Width

					if gameMap.CollisionLayer[x][y] {
						// If this tile is collision, no need to check again.
						continue
					}

					gameMap.CollisionLayer[x][y] = data[i] != 0
				}
			} else if layerType == "objectgroup" {
				objects, ok := mapLayer["objects"].([]any)
				if !ok {
					return nil, fmt.Errorf("invalid collision layer objects %s", name)
				}

				for _, object := range objects {
					objMap, ok := object.(map[string]any)
					if !ok {
						return nil, fmt.Errorf("invalid object in collision %s", name)
					}

					objectHeight, ok := objMap["height"].(float64)
					if !ok {
						return nil, fmt.Errorf("invalid object height of %s", name)
					}

					objectWidth, ok := objMap["width"].(float64)
					if !ok {
						return nil, fmt.Errorf("invalid object width of %s", name)
					}

					if gameMap.TileHeight != int(objectHeight) || gameMap.TileWidth != int(objectWidth) {
						return nil, fmt.Errorf(
							"object size of collision layer %s to be different from tile size", name)
					}

					xPixel, ok := objMap["x"].(float64)
					if !ok {
						return nil, fmt.Errorf("invalid object x of %s", name)
					}

					yPixel, ok := objMap["y"].(float64)
					if !ok {
						return nil, fmt.Errorf("invalid object y of %s", name)
					}

					xTile := int(xPixel) / gameMap.TileWidth
					yTile := int(yPixel) / gameMap.TileHeight

					gameMap.CollisionLayer[xTile][yTile] = true
				}
			} else {
				return nil, fmt.Errorf("invalid layer type %s of %s", layerType, name)
			}
		}
	}

	if len(collisionLayers) > 0 {
		return nil, fmt.Errorf("not found collision layer %v", collisionLayers)
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

type Message struct {
	UserID    string    `json:"user_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
