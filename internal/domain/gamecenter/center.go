package gamecenter

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type EngineInfo struct {
	roomIDs  []string
	lastPing time.Time
}

type GameCenter struct {
	mutex          sync.Mutex
	pendingRoomIDs []string
	engines        map[string]*EngineInfo

	gameRepo          repository.GameRepository
	gameCharacterRepo repository.GameCharacterRepository
	communityRepo     repository.CommunityRepository
	publisher         pubsub.Publisher
	storage           storage.Storage
}

func NewGameCenter(
	gameRepo repository.GameRepository,
	gameCharacterRepo repository.GameCharacterRepository,
	communityRepo repository.CommunityRepository,
	publisher pubsub.Publisher,
	storage storage.Storage,
) *GameCenter {
	return &GameCenter{
		mutex:             sync.Mutex{},
		pendingRoomIDs:    make([]string, 0),
		engines:           make(map[string]*EngineInfo),
		gameRepo:          gameRepo,
		gameCharacterRepo: gameCharacterRepo,
		communityRepo:     communityRepo,
		publisher:         publisher,
		storage:           storage,
	}
}

func (gc *GameCenter) Init(ctx context.Context) error {
	rooms, err := gc.gameRepo.GetAllRooms(ctx)
	if err != nil {
		return err
	}

	for _, room := range rooms {
		if room.StartedBy == "" {
			gc.pendingRoomIDs = append(gc.pendingRoomIDs, room.ID)
		} else {
			if _, ok := gc.engines[room.StartedBy]; !ok {
				gc.engines[room.StartedBy] = &EngineInfo{roomIDs: nil, lastPing: time.Now()}
			}

			gc.engines[room.StartedBy].roomIDs = append(gc.engines[room.StartedBy].roomIDs, room.ID)
		}
	}

	// If a community has no room, we will create one.
	communitiesWithNoGame, err := gc.communityRepo.GetWithNoGameRoom(ctx)
	if err != nil {
		return err
	}

	for _, community := range communitiesWithNoGame {
		gc.handleCreateRoom(ctx, community.ID)
	}

	return nil
}

func (gc *GameCenter) HandleEvent(ctx context.Context, topic string, pack *pubsub.Pack, tt time.Time) {
	switch topic {
	case model.GameEnginePingTopic:
		gc.handlePing(ctx, string(pack.Key))

	case model.CreateRoomTopic:
		gc.handleCreateRoom(ctx, string(pack.Key))

	case model.CreateCharacterTopic:
		gc.handleCreateCharacter(ctx, string(pack.Key))
	}
}

func (gc *GameCenter) handlePing(ctx context.Context, engineID string) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	if _, ok := gc.engines[engineID]; !ok {
		gc.engines[engineID] = &EngineInfo{}
	}
	gc.engines[engineID].lastPing = time.Now()
}

func (gc *GameCenter) handleCreateRoom(ctx context.Context, roomID string) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	gc.pendingRoomIDs = append(gc.pendingRoomIDs, roomID)
}

func (gc *GameCenter) handleCreateCharacter(ctx context.Context, characterID string) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	character, err := gc.gameCharacterRepo.GetByID(ctx, characterID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get character: %v", err)
		return
	}

	characterData, err := gc.storage.DownloadFromURL(ctx, character.ConfigURL)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot download character data: %v", err)
		return
	}

	parsedCharacter, err := gameengine.ParseCharacter(characterData)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot parse character: %v", err)
		return
	}

	engineCharacter := gameengine.Character{
		ID:    character.ID,
		Name:  character.Name,
		Level: character.Level,
		Size: gameengine.Size{
			Width:  parsedCharacter.Width,
			Height: parsedCharacter.Height,
			Sprite: gameengine.Sprite{
				WidthRatio:  character.SpriteWidthRatio,
				HeightRatio: character.SpriteHeightRatio,
			},
		},
	}

	serverAction := []map[string]any{{
		"user_id": "",
		"type":    gameengine.CreateCharacterAction{}.Type(),
		"value":   engineCharacter,
	}}

	b, err := json.Marshal(serverAction)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot marshal action: %v", err)
		return
	}

	for engineID := range gc.engines {
		err := gc.publisher.Publish(ctx, engineID, &pubsub.Pack{
			Key: []byte{},
			Msg: b,
		})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot publish the create character topic: %v", err)
			return
		}
	}

	xcontext.Logger(ctx).Infof("Broadcast create character completed")
}

// Janitor removes game engines which not ping to game center for a long time.
func (gc *GameCenter) Janitor(ctx context.Context) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	xcontext.Logger(ctx).Infof("Janitor started")
	defer xcontext.Logger(ctx).Infof("Janitor completed")

	defer time.AfterFunc(xcontext.Configs(ctx).Game.GameCenterJanitorFrequency, func() {
		gc.Janitor(ctx)
	})

	for id, engine := range gc.engines {
		if time.Since(engine.lastPing) > xcontext.Configs(ctx).Game.GameCenterJanitorFrequency {
			for _, roomID := range engine.roomIDs {
				if err := gc.gameRepo.UpdateRoomEngine(ctx, roomID, ""); err != nil {
					xcontext.Logger(ctx).Errorf("Cannot empty room engine id: %v", err)
					continue
				}

				gc.pendingRoomIDs = append(gc.pendingRoomIDs, roomID)
			}

			delete(gc.engines, id)
			xcontext.Logger(ctx).Infof("Removed engine %s", id)
		}
	}
}

// LoadBalance navigates pending rooms to suitable game engine.
func (gc *GameCenter) LoadBalance(ctx context.Context) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	xcontext.Logger(ctx).Infof("Load balance started, len(queue)=%d", len(gc.pendingRoomIDs))
	defer xcontext.Logger(ctx).Infof("Load balance completed")

	defer time.AfterFunc(xcontext.Configs(ctx).Game.GameCenterLoadBalanceFrequency, func() {
		gc.LoadBalance(ctx)
	})

	for len(gc.pendingRoomIDs) > 0 {
		roomID := gc.pendingRoomIDs[0]
		engineID := gc.getTheMostIdleEngine(ctx)
		if engineID == "" {
			xcontext.Logger(ctx).Errorf("Cannot find any engine for load balance")
			return
		}

		if err := gc.publisher.Publish(ctx, engineID, &pubsub.Pack{Key: []byte(roomID)}); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot publish the control topic: %v", err)
			return
		}

		if err := gc.gameRepo.UpdateRoomEngine(ctx, roomID, engineID); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot update room engine id: %v", err)
			return
		}

		gc.engines[engineID].roomIDs = append(gc.engines[engineID].roomIDs, roomID)
		gc.pendingRoomIDs = gc.pendingRoomIDs[1:]
	}
}

func (gc *GameCenter) getTheMostIdleEngine(ctx context.Context) string {
	minRooms := -1
	suitableEngineID := ""
	for engineID := range gc.engines {
		if suitableEngineID == "" || len(gc.engines[engineID].roomIDs) < minRooms {
			minRooms = len(gc.engines[engineID].roomIDs)
			suitableEngineID = engineID
		}
	}

	return suitableEngineID
}
