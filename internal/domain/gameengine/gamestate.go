package gameengine

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/domain/statistic"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type GameState struct {
	roomID      string
	communityID string

	mapConfig  *GameMap
	characters []*Character

	// Initial position if user hadn't joined the room yet.
	initCenterPixelPosition Position

	// userDiff contains all user differences between the original game state vs
	// the current game state.
	// DO NOT modify this field directly, please use setter methods instead.
	userDiff *xsync.MapOf[string, *entity.GameUser]

	// userMap contains user information in this game. It uses pixel unit to
	// determine its position.
	userMap map[string]*User

	gameRepo          repository.GameRepository
	gameLuckyboxRepo  repository.GameLuckyboxRepository
	gameCharacterRepo repository.GameCharacterRepository
	userRepo          repository.UserRepository
	followerRepo      repository.FollowerRepository
	leaderboard       statistic.Leaderboard
	storage           storage.Storage

	// actionDelay indicates how long the action can be applied again.
	actionDelay map[string]time.Duration

	// messageHistory stores last messages of game.
	messageHistory []Message

	// luckybox information.
	luckyboxes               map[string]Luckybox
	luckyboxesByTilePosition map[Position]Luckybox

	// luckyboxDiff contains all luckybox differences between the original game
	// state vs the current game state.
	// DO NOT modify this field directly, please use setter methods instead.
	luckyboxDiff *xsync.MapOf[string, *entity.GameLuckybox]
}

// newGameState creates a game state given a room id.
func newGameState(
	ctx context.Context,
	gameRepo repository.GameRepository,
	gameLuckyboxRepo repository.GameLuckyboxRepository,
	gameCharacterRepo repository.GameCharacterRepository,
	userRepo repository.UserRepository,
	followerRepo repository.FollowerRepository,
	leaderboard statistic.Leaderboard,
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

	configData, err := storage.DownloadFromURL(ctx, gameMap.ConfigURL)
	if err != nil {
		return nil, err
	}

	var config MapConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, err
	}

	mapData, err := storage.DownloadFromURL(ctx, config.PathOf(config.Config))
	if err != nil {
		return nil, err
	}

	// Parse tmx map content from game map.
	parsedMap, err := ParseGameMap(mapData, config.CollisionLayers)
	if err != nil {
		return nil, err
	}

	characters, err := gameCharacterRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var characterList []*Character
	for _, character := range characters {
		characterData, err := storage.DownloadFromURL(ctx, character.ConfigURL)
		if err != nil {
			return nil, err
		}

		parsedCharacter, err := ParseCharacter(characterData)
		if err != nil {
			return nil, err
		}

		characterList = append(characterList, &Character{
			ID:    character.ID,
			Name:  character.Name,
			Level: character.Level,
			Size: Size{
				Width:  parsedCharacter.Width,
				Height: parsedCharacter.Height,
				Sprite: Sprite{
					WidthRatio:  character.SpriteWidthRatio,
					HeightRatio: character.SpriteHeightRatio,
				},
			},
		})
	}

	gameCfg := xcontext.Configs(ctx).Game
	gamestate := &GameState{
		roomID:                  room.ID,
		communityID:             room.CommunityID,
		mapConfig:               parsedMap,
		characters:              characterList,
		userDiff:                xsync.NewMapOf[*entity.GameUser](),
		luckyboxDiff:            xsync.NewMapOf[*entity.GameLuckybox](),
		gameRepo:                gameRepo,
		gameLuckyboxRepo:        gameLuckyboxRepo,
		gameCharacterRepo:       gameCharacterRepo,
		userRepo:                userRepo,
		followerRepo:            followerRepo,
		leaderboard:             leaderboard,
		storage:                 storage,
		messageHistory:          make([]Message, 0, gameCfg.MessageHistoryLength),
		initCenterPixelPosition: config.InitPosition,
		actionDelay: map[string]time.Duration{
			InitAction{}.Type():            gameCfg.InitActionDelay,
			JoinAction{}.Type():            gameCfg.JoinActionDelay,
			MessageAction{}.Type():         gameCfg.MessageActionDelay,
			CollectLuckyboxAction{}.Type(): gameCfg.CollectLuckyboxActionDelay,
		},
	}

	for _, character := range characterList {
		topLeftInitPosition := gamestate.initCenterPixelPosition.CenterToTopLeft(character.Size)
		if gamestate.mapConfig.IsCollision(topLeftInitPosition, character.Size) {
			return nil, fmt.Errorf("initial of character %s is standing on a collision object",
				character.Name)
		}
	}

	initPositionInTile := gamestate.mapConfig.pixelToTile(gamestate.initCenterPixelPosition)
	gamestate.mapConfig.CalculateReachableTileMap(initPositionInTile)

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
		if _, err := g.addUserToGame(ctx, gameUser); err != nil {
			return err
		}
	}

	return nil
}

// LoadLuckybox loads all available luckyboxes into game state.
func (g *GameState) LoadLuckybox(ctx context.Context) error {
	luckyboxes, err := g.gameLuckyboxRepo.GetAvailableLuckyboxesByRoomID(ctx, g.roomID)
	if err != nil {
		return err
	}

	g.luckyboxes = make(map[string]Luckybox)
	g.luckyboxesByTilePosition = make(map[Position]Luckybox)
	for _, luckybox := range luckyboxes {
		luckyboxState := Luckybox{
			ID:      luckybox.ID,
			EventID: luckybox.EventID,
			Point:   luckybox.Point,
			PixelPosition: Position{
				X: luckybox.PositionX,
				Y: luckybox.PositionY,
			},
		}

		luckyboxTilePosition := g.mapConfig.pixelToTile(luckyboxState.PixelPosition)

		if _, ok := g.mapConfig.CollisionTileMap[luckyboxTilePosition]; ok {
			xcontext.Logger(ctx).Errorf("Luckybox %s appears on collision layer", luckyboxState.ID)
			continue
		}

		if another, ok := g.luckyboxesByTilePosition[luckyboxTilePosition]; ok {
			xcontext.Logger(ctx).Errorf("Luckybox %s overlaps on %s", luckyboxState.ID, another.ID)
			continue
		}

		g.addLuckybox(luckyboxState)
	}

	return nil
}

// Apply applies an action into game state.
func (g *GameState) Apply(
	ctx context.Context, proxyID string, action Action,
) ([]model.GameActionServerRequest, error) {
	if delay, ok := g.actionDelay[action.Type()]; ok {
		if user, ok := g.userMap[action.Owner()]; ok {
			if last, ok := user.LastTimeAction[action.Type()]; ok && time.Since(last) < delay {
				return nil, fmt.Errorf("submit action %s too fast", action.Type())
			}
		}
	}

	replyActions, err := action.Apply(ctx, proxyID, g)
	if err != nil {
		return nil, err
	}

	if user, ok := g.userMap[action.Owner()]; ok {
		user.LastTimeAction[action.Type()] = time.Now()
	}

	return replyActions, nil
}

// SerializeUser returns a bytes object in JSON format representing for current
// position of all users.
func (g *GameState) SerializeUser() []User {
	var users []User
	for _, user := range g.userMap {
		if user.ConnectedBy.Valid {
			clientUser := *user
			clientUser.PixelPosition = clientUser.PixelPosition.TopLeftToCenter(user.Character.Size)
			users = append(users, clientUser)
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

// LuckyboxDiff returns all database tracking differences of game luckybox until
// now. The diff will be reset after this method is called.
//
// Usage example:
//
//   for _, luckybox := range gamestate.LuckyboxDiff() {
//       gameRepo.UpdateLuckybox(ctx, luckybox)
//   }
func (g *GameState) LuckyboxDiff() []*entity.GameLuckybox {
	diff := []*entity.GameLuckybox{}
	g.luckyboxDiff.Range(func(key string, value *entity.GameLuckybox) bool {
		diff = append(diff, value)
		g.luckyboxDiff.Delete(key)
		return true
	})

	return diff
}

// trackUserPosition tracks the position of user to update in database.
func (g *GameState) trackUserPosition(userID string, direction entity.DirectionType, position Position) {
	diff := g.loadOrStoreUserDiff(userID)
	if diff == nil {
		return
	}

	diff.PositionX = position.X
	diff.PositionY = position.Y
	diff.Direction = direction

	g.userMap[userID].PixelPosition = position
	g.userMap[userID].Direction = direction
}

// trackUserProxy tracks the proxy user is connecting.
func (g *GameState) trackUserProxy(userID string, proxyID string) {
	diff := g.loadOrStoreUserDiff(userID)
	if diff == nil {
		return
	}

	connectedBy := sql.NullString{Valid: false}
	if proxyID != "" {
		connectedBy = sql.NullString{Valid: true, String: proxyID}
	}

	diff.ConnectedBy = connectedBy
	g.userMap[userID].ConnectedBy = connectedBy
}

// trackUserCharacter tracks the character of user to update in database.
func (g *GameState) trackUserCharacter(userID string, character *Character) {
	diff := g.loadOrStoreUserDiff(userID)
	if diff == nil {
		return
	}

	diff.CharacterID = sql.NullString{Valid: true, String: character.ID}
	g.userMap[userID].Character = character
}

// trackNewUserCharacter tracks the new character of user.
func (g *GameState) trackNewUserCharacter(userID string, character *Character) {
	user, ok := g.userMap[userID]
	if !ok {
		return
	}

	user.OwnedCharacters = append(user.OwnedCharacters, character)
}

func (g *GameState) loadOrStoreUserDiff(userID string) *entity.GameUser {
	user, ok := g.userMap[userID]
	if !ok {
		return nil
	}

	gameUser, _ := g.userDiff.LoadOrStore(user.User.ID, &entity.GameUser{
		UserID:      user.User.ID,
		RoomID:      g.roomID,
		CharacterID: sql.NullString{Valid: true, String: user.Character.ID},
		PositionX:   user.PixelPosition.X,
		PositionY:   user.PixelPosition.Y,
		Direction:   user.Direction,
		ConnectedBy: user.ConnectedBy,
	})

	return gameUser
}

// addUser creates a new user in room.
func (g *GameState) addUser(user User) {
	g.userDiff.Store(user.User.ID, &entity.GameUser{
		UserID:      user.User.ID,
		RoomID:      g.roomID,
		CharacterID: sql.NullString{Valid: true, String: user.Character.ID},
		PositionX:   user.PixelPosition.X,
		PositionY:   user.PixelPosition.Y,
		Direction:   user.Direction,
		ConnectedBy: user.ConnectedBy,
	})

	g.userMap[user.User.ID] = &user
}

// removeLuckybox marks the luckybox as collected.
func (g *GameState) removeLuckybox(luckyboxID string, userID string) {
	luckybox, ok := g.luckyboxes[luckyboxID]
	if !ok {
		return
	}

	delete(g.luckyboxes, luckyboxID)
	delete(g.luckyboxesByTilePosition, g.mapConfig.pixelToTile(luckybox.PixelPosition))

	collectedBy := sql.NullString{Valid: false}
	collectedAt := sql.NullTime{Valid: false}
	if userID != "" {
		collectedBy = sql.NullString{Valid: true, String: userID}
		collectedAt = sql.NullTime{Valid: true, Time: time.Now()}
	}

	g.luckyboxDiff.Store(luckybox.ID, &entity.GameLuckybox{
		Base:        entity.Base{ID: luckybox.ID},
		EventID:     luckybox.EventID,
		PositionX:   luckybox.PixelPosition.X,
		PositionY:   luckybox.PixelPosition.Y,
		Point:       luckybox.Point,
		CollectedBy: collectedBy,
		CollectedAt: collectedAt,
	})
}

// addLuckybox creates a new luckybox in room.
func (g *GameState) addLuckybox(luckybox Luckybox) {
	g.luckyboxDiff.Store(luckybox.ID, &entity.GameLuckybox{
		Base:        entity.Base{ID: luckybox.ID},
		EventID:     luckybox.EventID,
		PositionX:   luckybox.PixelPosition.X,
		PositionY:   luckybox.PixelPosition.Y,
		Point:       luckybox.Point,
		CollectedBy: sql.NullString{},
		CollectedAt: sql.NullTime{},
	})

	g.luckyboxes[luckybox.ID] = luckybox
	g.luckyboxesByTilePosition[g.mapConfig.pixelToTile(luckybox.PixelPosition)] = luckybox
}

func (g *GameState) findCharacterByID(id string) *Character {
	for _, p := range g.characters {
		if p.ID == id {
			return p
		}
	}

	return nil
}

func (g *GameState) addUserToGame(ctx context.Context, gameUser entity.GameUser) (bool, error) {
	userCharacters, err := g.gameCharacterRepo.GetAllUserCharacters(
		ctx, gameUser.UserID, g.communityID)
	if err != nil {
		return false, err
	}

	if len(userCharacters) == 0 {
		return false, nil
	}

	if !gameUser.CharacterID.Valid {
		gameUser.CharacterID = sql.NullString{Valid: true, String: userCharacters[0].CharacterID}
	}

	character := g.findCharacterByID(gameUser.CharacterID.String)
	if character == nil {
		xcontext.Logger(ctx).Errorf("Not found character %s of user %s",
			gameUser.CharacterID, gameUser.UserID)
		return false, nil
	}

	userPixelPosition := Position{X: gameUser.PositionX, Y: gameUser.PositionY}
	if g.mapConfig.IsCollision(userPixelPosition, character.Size) {
		xcontext.Logger(ctx).Errorf("Detected a user standing on a collision tile at pixel %s", userPixelPosition)
		return false, nil
	}

	user, err := g.userRepo.GetByID(ctx, gameUser.UserID)
	if err != nil {
		return false, err
	}

	g.addUser(User{
		User: UserInfo{
			ID:        user.ID,
			Name:      user.Name,
			AvatarURL: user.ProfilePicture,
		},
		Character:      character,
		Direction:      gameUser.Direction,
		PixelPosition:  userPixelPosition,
		LastTimeAction: make(map[string]time.Time),
		ConnectedBy:    gameUser.ConnectedBy,
	})

	for _, uc := range userCharacters {
		character := g.findCharacterByID(uc.CharacterID)
		if character == nil {
			xcontext.Logger(ctx).Warnf("Cannot found character %s of user %s",
				uc.CharacterID, gameUser.UserID)
			continue
		}

		g.trackNewUserCharacter(gameUser.UserID, character)
	}

	return true, nil
}
