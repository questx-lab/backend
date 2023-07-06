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

func (gc *GameCenter) Ping(ctx context.Context, domainName string) error {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	engineIP := domainName
	if engineIP == "" {
		engineIP, _, _ = strings.Cut(rpc.PeerInfoFromContext(ctx).RemoteAddr, ":")
	}

	if engineIP == "" {
		xcontext.Logger(gc.rootCtx).Errorf("Not found remote address or domain name")
		return errors.New("not found remote address or domain name")
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
			for _, roomID := range engine.roomIDs {
				if err := gc.gameRepo.UpdateRoomEngine(ctx, roomID, ""); err != nil {
					xcontext.Logger(ctx).Errorf("Cannot empty room engine id: %v", err)
					continue
				}

				gc.pendingRoomIDs = append(gc.pendingRoomIDs, roomID)
			}

			engine.caller.Close()
			delete(gc.engines, ip)
			xcontext.Logger(ctx).Infof("Removed engine %s", ip)
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
