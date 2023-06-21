package gamecenter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/pubsub"
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

	gameRepo      repository.GameRepository
	communityRepo repository.CommunityRepository
	publisher     pubsub.Publisher
}

func NewGameCenter(
	gameRepo repository.GameRepository,
	communityRepo repository.CommunityRepository,
	publisher pubsub.Publisher,
) *GameCenter {
	return &GameCenter{
		mutex:          sync.Mutex{},
		pendingRoomIDs: make([]string, 0),
		engines:        make(map[string]*EngineInfo),
		gameRepo:       gameRepo,
		communityRepo:  communityRepo,
		publisher:      publisher,
	}
}

func (gc *GameCenter) Init(ctx context.Context) error {
	rooms, err := gc.gameRepo.GetAllRooms(ctx)
	if err != nil {
		return err
	}

	for _, room := range rooms {
		if room.StartedBy == "" {
			gc.pendingRoomIDs = append(gc.pendingRoomIDs, room.StartedBy)
		} else {
			if _, ok := gc.engines[room.StartedBy]; !ok {
				gc.engines[room.StartedBy] = &EngineInfo{roomIDs: nil, lastPing: time.Now()}
			}

			gc.engines[room.StartedBy].roomIDs = append(gc.engines[room.StartedBy].roomIDs, room.ID)
		}
	}

	return nil
}

func (gc *GameCenter) HandleEvent(ctx context.Context, topic string, pack *pubsub.Pack, tt time.Time) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	switch topic {
	case model.GameEnginePingTopic:
		gc.handlePing(ctx, string(pack.Key))

	case model.CreateCommunityTopic:
		gc.handleCreateRoom(ctx, string(pack.Key))
	}
}

func (gc *GameCenter) handlePing(ctx context.Context, engineID string) {
	if _, ok := gc.engines[engineID]; !ok {
		gc.engines[engineID] = &EngineInfo{}
	}
	gc.engines[engineID].lastPing = time.Now()
}

func (gc *GameCenter) handleCreateRoom(ctx context.Context, communityID string) {
	community, err := gc.communityRepo.GetByID(ctx, communityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Not found community id %s: %v", communityID, err)
		return
	}

	firstMap, err := gc.gameRepo.GetFirstMap(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Not found the first map in db: %v", err)
		return
	}

	room := entity.GameRoom{
		Base:        entity.Base{ID: uuid.NewString()},
		CommunityID: communityID,
		MapID:       firstMap.ID,
		Name:        fmt.Sprintf("%s-%d", community.Handle, crypto.RandRange(100, 999)),
	}
	if err := gc.gameRepo.CreateRoom(ctx, &room); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create room for %s: %v", community.Handle, err)
		return
	}

	xcontext.Logger(ctx).Infof("Create room %s successfully", room.Name)
	gc.pendingRoomIDs = append(gc.pendingRoomIDs, room.ID)
}

// Janitor removes game engines which not ping to game center for a long time.
func (gc *GameCenter) Janitor(ctx context.Context) {
	xcontext.Logger(ctx).Infof("Janitor started")
	defer xcontext.Logger(ctx).Infof("Janitor completed")

	gc.mutex.Lock()
	defer gc.mutex.Unlock()
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
