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

type CharacterConfig struct {
	Name   string `json:"name"`
	Config string `json:"config"`
}

type MapConfig struct {
	BaseURL          string            `json:"base_url"`
	Config           string            `json:"config"`
	CharacterConfigs []CharacterConfig `json:"character_configs"`
	CollisionLayers  []string          `json:"collision_layers"`
	InitPosition     Position          `json:"init_position"`
}

func (c MapConfig) PathOf(path string) string {
	return c.BaseURL + path
}

type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Character struct {
	Name string `json:"name"`
	Size Size   `json:"-"`
}

type UserInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type User struct {
	User UserInfo `json:"user"`

	// Character specifies the character which this user is using.
	Character Character `json:"player"` // TODO: For back-compatible.

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

func (p Position) CenterToTopLeft(s Size) Position {
	return Position{p.X - s.Width/2, p.Y - s.Height/2}
}

func (p Position) TopLeftToCenter(s Size) Position {
	return Position{p.X + s.Width/2, p.Y + s.Height/2}
}

type GameMap struct {
	// The map size in number of tiles.
	MapSizeInTile Size

	// Tile size.
	TileSizeInPixel Size

	CollisionTileMap map[Position]any
	ReachableTileMap map[Position]any
}

func (g *GameMap) CalculateReachableTileMap(initPositionInTile Position) {
	pendingPositions := []Position{initPositionInTile}
	pendingPositionsMap := map[Position]any{
		initPositionInTile: nil,
	}

	for len(pendingPositions) > 0 {
		currentPosition := pendingPositions[0]
		g.ReachableTileMap[currentPosition] = nil

		availablePositions := []Position{
			{X: currentPosition.X + 1, Y: currentPosition.Y}, // right
			{X: currentPosition.X - 1, Y: currentPosition.Y}, // left
			{X: currentPosition.X, Y: currentPosition.Y + 1}, // down
			{X: currentPosition.X, Y: currentPosition.Y - 1}, // up
		}

		for _, pos := range availablePositions {
			if pos.X < 0 || pos.X >= g.MapSizeInTile.Width {
				continue
			}

			if pos.Y < 0 || pos.Y >= g.MapSizeInTile.Height {
				continue
			}

			if _, ok := pendingPositionsMap[pos]; ok {
				continue
			}

			if _, ok := g.CollisionTileMap[pos]; ok {
				continue
			}

			pendingPositions = append(pendingPositions, pos)
			pendingPositionsMap[pos] = nil
		}

		pendingPositions = pendingPositions[1:]
	}
}

// IsCollision checks if the object is collided with any collision tile or
// not. The object is represented by its top left point, width, and height. All
// parameters must be in pixel.
func (g *GameMap) IsCollision(topLeftInPixel Position, size Size) bool {
	if g.IsPointCollision(topLeftInPixel) {
		return true
	}

	if g.IsPointCollision(topRight(topLeftInPixel, size)) {
		return true
	}

	if g.IsPointCollision(bottomLeft(topLeftInPixel, size)) {
		return true
	}

	if g.IsPointCollision(bottomRight(topLeftInPixel, size)) {
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

	if tilePosition.X >= g.MapSizeInTile.Width || tilePosition.Y >= g.MapSizeInTile.Height {
		return true
	}

	return false
}

// pixelToTile returns position in tile given a position in pixel.
func (g *GameMap) pixelToTile(p Position) Position {
	return Position{X: p.X / g.TileSizeInPixel.Width, Y: p.Y / g.TileSizeInPixel.Height}
}

// tileToPixel returns the top left position of a tile in pixel.
func (g *GameMap) tileToPixel(p Position) Position {
	return Position{
		X: p.X * g.TileSizeInPixel.Width,
		Y: p.Y * g.TileSizeInPixel.Height,
	}
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
		MapSizeInTile: Size{
			Height: int(height),
			Width:  int(width),
		},
		TileSizeInPixel: Size{
			Height: int(tileHeight),
			Width:  int(tileWidth),
		},
		CollisionTileMap: make(map[Position]any),
		ReachableTileMap: make(map[Position]any),
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

				if len(data) != gameMap.MapSizeInTile.Width*gameMap.MapSizeInTile.Height {
					return nil, fmt.Errorf("invalid number of elements in collision layer data %s", name)
				}

				for i := range data {
					if data[i] == 0.0 {
						continue
					}

					pos := Position{
						X: i % gameMap.MapSizeInTile.Width,
						Y: i / gameMap.MapSizeInTile.Width,
					}
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

					if gameMap.TileSizeInPixel.Height != int(objectHeight) || gameMap.TileSizeInPixel.Width != int(objectWidth) {
						return nil, fmt.Errorf(
							"object size of collision layer %s is different from tile size", name)
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
						X: int(xPixel) / gameMap.TileSizeInPixel.Width,
						Y: int(yPixel) / gameMap.TileSizeInPixel.Height,
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

type GameCharacter struct {
	Height int
	Width  int
}

func ParseCharacter(jsonContent []byte) (*GameCharacter, error) {
	m := map[string]any{}
	err := json.Unmarshal(jsonContent, &m)
	if err != nil {
		return nil, err
	}

	frameMap, ok := m["frames"].(map[string]any)
	if ok {
		for _, frame := range frameMap {
			return parseCharacterFrame(frame)
		}
	}

	frameArr, ok := m["frames"].([]any)
	if ok {
		if len(frameArr) == 0 {
			return nil, errors.New("not found any frames")
		}

		return parseCharacterFrame(frameArr[0])
	}

	return nil, errors.New("invalid or not found frames in character")
}

func parseCharacterFrame(frame any) (*GameCharacter, error) {
	frameValue, ok := frame.(map[string]any)
	if !ok {
		return nil, errors.New("invalid character")
	}

	sourceSize, ok := frameValue["sourceSize"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid character source size")
	}

	w, ok := sourceSize["w"].(float64)
	if !ok {
		return nil, errors.New("invalid width")
	}

	h, ok := sourceSize["h"].(float64)
	if !ok {
		return nil, errors.New("invalid height")
	}

	return &GameCharacter{Height: int(h), Width: int(w)}, nil
}

type Message struct {
	User      UserInfo  `json:"user"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type Luckybox struct {
	ID      string `json:"id"`
	EventID string `json:"event_id"`
	Point   int    `json:"point"`

	// Position of luckybox in tile.
	PixelPosition Position `json:"position"`
}

func (b Luckybox) WithCenterPixelPosition(tileSize Size) Luckybox {
	newbox := b
	newbox.PixelPosition = newbox.PixelPosition.TopLeftToCenter(tileSize)
	return newbox
}
