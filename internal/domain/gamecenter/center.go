package gamecenter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type EngineInfo struct {
	roomIDs  []string
	lastPing time.Time
	caller   client.GameEngineCaller
}

type GameCenter struct {
	rootCtx context.Context

	mutex          sync.Mutex
	pendingRoomIDs []string
	engines        map[string]*EngineInfo

	gameRepo          repository.GameRepository
	gameLuckyboxRepo  repository.GameLuckyboxRepository
	gameCharacterRepo repository.GameCharacterRepository
	communityRepo     repository.CommunityRepository
	storage           storage.Storage
}

func NewGameCenter(
	ctx context.Context,
	gameRepo repository.GameRepository,
	gameLuckyboxRepo repository.GameLuckyboxRepository,
	gameCharacterRepo repository.GameCharacterRepository,
	communityRepo repository.CommunityRepository,
	storage storage.Storage,
) *GameCenter {
	return &GameCenter{
		rootCtx:           ctx,
		mutex:             sync.Mutex{},
		pendingRoomIDs:    make([]string, 0),
		engines:           make(map[string]*EngineInfo),
		gameRepo:          gameRepo,
		gameLuckyboxRepo:  gameLuckyboxRepo,
		gameCharacterRepo: gameCharacterRepo,
		communityRepo:     communityRepo,
		storage:           storage,
	}
}

func (gc *GameCenter) Init(ctx context.Context) error {
	rooms, err := gc.gameRepo.GetAllRooms(ctx)
	if err != nil {
		return err
	}

	for _, room := range rooms {
		if room.StartedBy != "" {
			if err := gc.refreshEngine(ctx, room.StartedBy); err != nil {
				xcontext.Logger(ctx).Warnf(
					"Cannot connect to engine %s of room %s: %v", room.StartedBy, room.ID, err)
			} else {
				gc.engines[room.StartedBy].roomIDs = append(gc.engines[room.StartedBy].roomIDs, room.ID)
				continue
			}
		}

		gc.pendingRoomIDs = append(gc.pendingRoomIDs, room.ID)
	}

	return nil
}

func (gc *GameCenter) Ping(_ctx context.Context, isNew bool) error {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	var engineIP string
	portIndex := strings.LastIndex(rpc.PeerInfoFromContext(_ctx).RemoteAddr, ":")
	if portIndex == -1 {
		engineIP = rpc.PeerInfoFromContext(_ctx).RemoteAddr
	} else {
		engineIP = rpc.PeerInfoFromContext(_ctx).RemoteAddr[:portIndex]
	}

	if engineIP == "[::1]" {
		engineIP = "127.0.0.1"
	}

	if engineIP == "" {
		xcontext.Logger(gc.rootCtx).Errorf("Not found remote address or domain name")
		return errors.New("not found remote address or domain name")
	}

	if isNew {
		gc.removeSingleEngine(gc.rootCtx, engineIP)
	}

	if err := gc.refreshEngine(gc.rootCtx, engineIP); err != nil {
		xcontext.Logger(gc.rootCtx).Errorf("Cannot refresh engine: %v", err)
		return err
	}

	gc.engines[engineIP].lastPing = time.Now()
	return nil
}

func (gc *GameCenter) StartRoom(_ context.Context, roomID string) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	xcontext.Logger(gc.rootCtx).Infof("Room %s is pending", roomID)
	gc.pendingRoomIDs = append(gc.pendingRoomIDs, roomID)
}

func (gc *GameCenter) CreateCharacter(_ context.Context, characterID string) error {
	gameCharacter, err := gc.gameCharacterRepo.GetByID(gc.rootCtx, characterID)
	if err != nil {
		xcontext.Logger(gc.rootCtx).Errorf("Cannot get character: %v", err)
		return err
	}

	characterData, err := gc.storage.DownloadFromURL(gc.rootCtx, gameCharacter.ConfigURL)
	if err != nil {
		xcontext.Logger(gc.rootCtx).Errorf("Cannot download character data: %v", err)
		return err
	}

	parsedCharacter, err := gameengine.ParseCharacter(characterData)
	if err != nil {
		xcontext.Logger(gc.rootCtx).Errorf("Cannot parse character: %v", err)
		return err
	}

	character := client.Character{
		ID:    gameCharacter.ID,
		Name:  gameCharacter.Name,
		Level: gameCharacter.Level,
		Size: client.Size{
			Width:  parsedCharacter.Width,
			Height: parsedCharacter.Height,
			Sprite: client.Sprite{
				WidthRatio:  gameCharacter.SpriteWidthRatio,
				HeightRatio: gameCharacter.SpriteHeightRatio,
			},
		},
	}

	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	for _, engine := range gc.engines {
		if err := engine.caller.CreateCharacter(gc.rootCtx, character); err != nil {
			return err
		}
	}

	return nil
}

func (gc *GameCenter) BuyCharacter(_ context.Context, userID, characterID, communityID string) error {
	gameRooms, err := gc.gameRepo.GetRoomsByUserCommunity(gc.rootCtx, userID, communityID)
	if err != nil {
		xcontext.Logger(gc.rootCtx).Errorf("Cannot get user by community: %v", err)
		return err
	}

	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	calledEngine := map[string]any{}
	for _, room := range gameRooms {
		if _, ok := calledEngine[room.StartedBy]; ok {
			continue
		}

		engine, ok := gc.engines[room.StartedBy]
		if !ok {
			continue
		}

		if err := engine.caller.BuyCharacter(gc.rootCtx, userID, characterID, communityID); err != nil {
			return err
		}

		calledEngine[room.StartedBy] = nil
	}

	return nil
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

	for ip, engine := range gc.engines {
		if time.Since(engine.lastPing) > xcontext.Configs(ctx).Game.GameCenterJanitorFrequency {
			gc.removeSingleEngine(ctx, ip)
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
		ip := gc.getTheMostIdleEngine(ctx)
		if ip == "" {
			xcontext.Logger(ctx).Errorf("Cannot find any engine for load balance")
			return
		}

		if err := gc.engines[ip].caller.StartRoom(ctx, roomID); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot call start room: %v", err)
			return
		}

		if err := gc.gameRepo.UpdateRoomEngine(ctx, roomID, ip); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot update room engine id: %v", err)
			return
		}

		gc.engines[ip].roomIDs = append(gc.engines[ip].roomIDs, roomID)
		gc.pendingRoomIDs = gc.pendingRoomIDs[1:]
	}
}

func (gc *GameCenter) ScheduleLuckyboxEvent(ctx context.Context) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	xcontext.Logger(ctx).Infof("Schedule luckybox event started")
	defer xcontext.Logger(ctx).Infof("Schedule luckybox event completed")

	defer time.AfterFunc(time.Until(time.Now().Add(time.Minute).Truncate(time.Minute)), func() {
		gc.ScheduleLuckyboxEvent(ctx)
	})

	// START EVENTS.
	shouldStartEvents, err := gc.gameLuckyboxRepo.GetShouldStartLuckyboxEvent(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get should-start events: %v", err)
		return
	}

	for _, event := range shouldStartEvents {
		room, err := gc.gameRepo.GetRoomByID(ctx, event.RoomID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get room: %v", err)
			continue
		}

		if room.StartedBy == "" {
			xcontext.Logger(ctx).Warnf("Game room has not started yet")
			continue
		}

		engine, ok := gc.engines[room.StartedBy]
		if !ok {
			xcontext.Logger(ctx).Warnf(
				"Not found engine of room %s when start luckybox %s", room.ID, event.ID)
			continue
		}

		if err := engine.caller.StartLuckyboxEvent(ctx, event.ID, event.RoomID); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot call start luckybox event %s: %v", event.ID, err)
			return
		}

		err = gc.gameLuckyboxRepo.MarkLuckyboxEventAsStarted(ctx, event.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot mark event %s as started: %v", event.ID, err)
			return
		}

		xcontext.Logger(ctx).Infof("Start event %s of room %s successfully", event.ID, room.ID)
	}

	// STOP EVENTS.
	shouldStopEvents, err := gc.gameLuckyboxRepo.GetShouldStopLuckyboxEvent(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get should-stop events: %v", err)
		return
	}

	for _, event := range shouldStopEvents {
		room, err := gc.gameRepo.GetRoomByID(ctx, event.RoomID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get room: %v", err)
			continue
		}

		if room.StartedBy == "" {
			xcontext.Logger(ctx).Errorf("Game room has not started yet")
			continue
		}

		engine, ok := gc.engines[room.StartedBy]
		if !ok {
			xcontext.Logger(ctx).Warnf(
				"Not found engine of room %s when stop luckybox %s", room.ID, event.ID)
			continue
		}

		if err := engine.caller.StopLuckyboxEvent(ctx, event.ID, event.RoomID); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot call stop luckybox event %s: %v", event.ID, err)
			return
		}

		err = gc.gameLuckyboxRepo.MarkLuckyboxEventAsStopped(ctx, event.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot mark event %s as stopped: %v", event.ID, err)
			return
		}

		xcontext.Logger(ctx).Infof("Stop event %s of room %s successfully", event.ID, room.ID)
	}
}

func (gc *GameCenter) refreshEngine(ctx context.Context, engineIP string) error {
	if _, ok := gc.engines[engineIP]; !ok {
		rpcIP := fmt.Sprintf("http://%s:%s",
			engineIP, xcontext.Configs(gc.rootCtx).GameEngineRPCServer.Port)

		rpcClient, err := rpc.DialContext(ctx, rpcIP)
		if err != nil {
			return err
		}

		xcontext.Logger(ctx).Infof("Add new game engine %s", engineIP)
		gc.engines[engineIP] = &EngineInfo{caller: client.NewGameEngineCaller(rpcClient)}
	}

	gc.engines[engineIP].lastPing = time.Now()
	return nil
}

func (gc *GameCenter) getTheMostIdleEngine(ctx context.Context) string {
	minRooms := -1
	suitableEngineIP := ""
	for engineID := range gc.engines {
		if suitableEngineIP == "" || len(gc.engines[engineID].roomIDs) < minRooms {
			minRooms = len(gc.engines[engineID].roomIDs)
			suitableEngineIP = engineID
		}
	}

	return suitableEngineIP
}

func (gc *GameCenter) removeSingleEngine(ctx context.Context, engineIP string) {
	engine, ok := gc.engines[engineIP]
	if !ok {
		return
	}

	for _, roomID := range engine.roomIDs {
		if err := gc.gameRepo.UpdateRoomEngine(ctx, roomID, ""); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot empty room engine id: %v", err)
			continue
		}

		gc.pendingRoomIDs = append(gc.pendingRoomIDs, roomID)
	}

	if engine.caller != nil {
		engine.caller.Close()
	}
	delete(gc.engines, engineIP)
	xcontext.Logger(ctx).Infof("Removed engine %s", engineIP)
}
