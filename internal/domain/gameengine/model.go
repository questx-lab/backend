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

type Player struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Width  int    `json:"-"`
	Height int    `json:"-"`
}

type UserInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type User struct {
	User UserInfo `json:"user"`

	// PlayerName specifies the player avatar name which this user is using.
	Player Player `json:"player"`

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

func (p Position) Distance(another Position) float64 {
	x2 := math.Pow(float64(p.X)-float64(another.X), 2)
	y2 := math.Pow(float64(p.Y)-float64(another.Y), 2)
	return math.Sqrt(x2 + y2)
}

func (p Position) CenterToTopLeft(player Player) Position {
	return Position{p.X - player.Width/2, p.Y - player.Height/2}
}

func (p Position) TopLeftToCenter(player Player) Position {
	return Position{p.X + player.Width/2, p.Y + player.Height/2}
}

type GameMap struct {
	Height           int
	Width            int
	TileHeight       int
	TileWidth        int
	CollisionTileMap map[Position]any
}

// IsPlayerCollision checks if the object is collided with any collision tile or
// not. The object is represented by its top left point, width, and height. All
// parameters must be in pixel.
func (g *GameMap) IsPlayerCollision(topLeftInPixel Position, player Player) bool {
	if g.IsPointCollision(topLeftInPixel) {
		return true
	}

	if g.IsPointCollision(topRight(topLeftInPixel, player.Width, player.Height)) {
		return true
	}

	if g.IsPointCollision(bottomLeft(topLeftInPixel, player.Width, player.Height)) {
		return true
	}

	if g.IsPointCollision(bottomRight(topLeftInPixel, player.Width, player.Height)) {
		return true
	}

	return false
}

// IsPointCollision checks if a point is collided with any collision tile or
// not. The point position must be in pixel.
func (g *GameMap) IsPointCollision(pointPixel Position) bool {
	if pointPixel.X < 0 || pointPixel.Y < 0 {
		return true
	}

	tilePosition := g.pixelToTile(pointPixel)
	_, isBlocked := g.CollisionTileMap[tilePosition]
	if isBlocked {
		return true
	}

	if tilePosition.X >= g.Width || tilePosition.Y >= g.Height {
		return true
	}

	return false
}

// pixelToTile returns position in tile given a position in pixel.
func (g *GameMap) pixelToTile(p Position) Position {
	return Position{X: p.X / g.TileWidth, Y: p.Y / g.TileHeight}
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

	gameMap.CollisionTileMap = make(map[Position]any, gameMap.Width)
	for i := range gameMap.CollisionTileMap {
		gameMap.CollisionTileMap[i] = make([]bool, gameMap.Height)
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
					if data[i] == 0.0 {
						continue
					}

					pos := Position{X: i % gameMap.Width, Y: i / gameMap.Width}
					if _, ok := gameMap.CollisionTileMap[pos]; !ok {
						gameMap.CollisionTileMap[pos] = nil
					}
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

					pos := Position{
						X: int(xPixel) / gameMap.TileWidth,
						Y: int(yPixel) / gameMap.TileHeight,
					}
					if _, ok := gameMap.CollisionTileMap[pos]; !ok {
						gameMap.CollisionTileMap[pos] = nil
					}
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

	frameMap, ok := m["frames"].(map[string]any)
	if ok {
		for _, frame := range frameMap {
			return parsePlayerFrame(frame)
		}
	}

	frameArr, ok := m["frames"].([]any)
	if ok {
		if len(frameArr) == 0 {
			return nil, errors.New("not found any frames")
		}

		return parsePlayerFrame(frameArr[0])
	}

	return nil, errors.New("invalid or not found frames in player")
}

func parsePlayerFrame(frame any) (*GamePlayer, error) {
	frameValue, ok := frame.(map[string]any)
	if !ok {
		return nil, errors.New("invalid player")
	}

	sourceSize, ok := frameValue["sourceSize"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid player source size")
	}

	w, ok := sourceSize["w"].(float64)
	if !ok {
		return nil, errors.New("invalid width")
	}

	h, ok := sourceSize["h"].(float64)
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
