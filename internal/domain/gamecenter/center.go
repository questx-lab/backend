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
	"github.com/questx-lab/backend/internal/repository"
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

	gameRepo      repository.GameRepository
	communityRepo repository.CommunityRepository
}

func NewGameCenter(
	ctx context.Context,
	gameRepo repository.GameRepository,
	communityRepo repository.CommunityRepository,
) *GameCenter {
	return &GameCenter{
		rootCtx:        ctx,
		mutex:          sync.Mutex{},
		pendingRoomIDs: make([]string, 0),
		engines:        make(map[string]*EngineInfo),
		gameRepo:       gameRepo,
		communityRepo:  communityRepo,
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

func (gc *GameCenter) Ping(_ctx context.Context, domainName string, isNew bool) error {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	engineIP := domainName
	if engineIP == "" {
		engineIP, _, _ = strings.Cut(rpc.PeerInfoFromContext(_ctx).RemoteAddr, ":")
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
	shouldStartEvents, err := gc.gameRepo.GetShouldStartLuckyboxEvent(ctx)
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

		err = gc.gameRepo.MarkLuckyboxEventAsStarted(ctx, event.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot mark event %s as started: %v", event.ID, err)
			return
		}

		xcontext.Logger(ctx).Infof("Start event %s of room %s successfully", event.ID, room.ID)
	}

	// STOP EVENTS.
	shouldStopEvents, err := gc.gameRepo.GetShouldStopLuckyboxEvent(ctx)
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

		err = gc.gameRepo.MarkLuckyboxEventAsStopped(ctx, event.ID)
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

	engine.caller.Close()
	delete(gc.engines, engineIP)
	xcontext.Logger(ctx).Infof("Removed engine %s", engineIP)
}
