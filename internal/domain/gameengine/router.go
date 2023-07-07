package gameengine

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/domain/statistic"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type Router interface {
	ID() string
	StartRoom(ctx context.Context, roomID string) error
	StopRoom(ctx context.Context, roomID string) error
	PingCenter(ctx context.Context, i int)
	ServeGameProxy(ctx context.Context, req *model.ServeGameProxyRequest) error
}

type router struct {
	rootCtx          context.Context
	hostname         string
	gameRepo         repository.GameRepository
	userRepo         repository.UserRepository
	followerRepo     repository.FollowerRepository
	leaderboard      statistic.Leaderboard
	storage          storage.Storage
	gameCenterCaller client.GameCenterCaller

	engines *xsync.MapOf[string, *engine]
}

func NewRouter(
	ctx context.Context,
	gameRepo repository.GameRepository,
	userRepo repository.UserRepository,
	followerRepo repository.FollowerRepository,
	leaderboard statistic.Leaderboard,
	storage storage.Storage,
	gameCenterClient *rpc.Client,
) Router {
	hostname := ""
	if xcontext.Configs(ctx).DomainNameSuffix != "" {
		hostname = os.Getenv("HOSTNAME") + xcontext.Configs(ctx).DomainNameSuffix
	}

	return &router{
		rootCtx:          ctx,
		hostname:         hostname,
		gameRepo:         gameRepo,
		userRepo:         userRepo,
		followerRepo:     followerRepo,
		leaderboard:      leaderboard,
		storage:          storage,
		gameCenterCaller: client.NewGameCenterCaller(gameCenterClient),
		engines:          xsync.NewMapOf[*engine](),
	}
}

func (r *router) ID() string {
	if r.hostname == "" {
		return os.Getenv("HOSTNAME")
	}

	return r.hostname
}

func (r *router) StartRoom(_ context.Context, roomID string) error {
	engine, err := NewEngine(r.rootCtx, r.gameRepo, r.userRepo, r.followerRepo,
		r.leaderboard, r.storage, roomID)
	if err != nil {
		xcontext.Logger(r.rootCtx).Errorf("Cannot start game %s: %v", roomID, err)
		return errorx.Unknown
	}

	r.engines.Store(roomID, engine)
	xcontext.Logger(r.rootCtx).Infof("Start game %s successfully", roomID)

	return nil
}

func (r *router) StopRoom(_ context.Context, roomID string) error {
	engine, ok := r.engines.LoadAndDelete(roomID)
	if !ok {
		return errorx.New(errorx.NotFound, "Not found room in this engine")
	}

	engine.Stop(r.rootCtx)
	xcontext.Logger(r.rootCtx).Infof("Stop game %s successfully", roomID)

	return nil
}

func (r *router) StartLuckyboxEvent(_ context.Context, eventID, roomID string) error {
	engine, ok := r.engines.Load(roomID)
	if !ok {
		return errorx.New(errorx.NotFound, "Not found room in start luckybox event")
	}

	engine.requestAction <- GameActionProxyRequest{
		ProxyID: "",
		Actions: []model.GameActionServerRequest{{
			UserID: "",
			Type:   StartLuckyboxEventAction{}.Type(),
			Value:  map[string]any{"event_id": eventID},
		}},
	}

	return nil
}

func (r *router) StopLuckyboxEvent(_ context.Context, eventID, roomID string) error {
	engine, ok := r.engines.Load(roomID)
	if !ok {
		return errorx.New(errorx.NotFound, "Not found room in stop luckybox event")
	}

	engine.requestAction <- GameActionProxyRequest{
		ProxyID: "",
		Actions: []model.GameActionServerRequest{{
			UserID: "",
			Type:   StopLuckyboxEventAction{}.Type(),
			Value:  map[string]any{"event_id": eventID},
		}},
	}

	return nil
}

func (r *router) PingCenter(ctx context.Context, i int) {
	nextIndex := i + 1
	if nextIndex == 0 {
		// Overflow detected. Never use index as 0 for nextIndex.
		nextIndex = 1
	}

	if err := r.gameCenterCaller.Ping(ctx, r.hostname, i == 0); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot ping center: %v", err)
		nextIndex = i
	}
	defer time.AfterFunc(xcontext.Configs(ctx).Game.GameEnginePingFrequency, func() {
		r.PingCenter(ctx, nextIndex)
	})

	if i%10 == 0 {
		xcontext.Logger(ctx).Infof("Ping center successfully")
	}
}

func (r *router) ServeGameProxy(ctx context.Context, req *model.ServeGameProxyRequest) error {
	engine, ok := r.engines.Load(req.RoomID)
	if !ok {
		return errorx.New(errorx.NotFound, "Not found room in this engine")
	}

	proxyID := uuid.NewString()
	responseAction := engine.RegisterProxy(ctx, proxyID)
	defer engine.UnregisterProxy(ctx, proxyID)

	wsClient := xcontext.WSClient(ctx)
	isStop := false
	for !isStop {
		select {
		case msg, ok := <-wsClient.R:
			if !ok {
				isStop = true
				break
			}

			serverAction := []model.GameActionServerRequest{}
			err := json.Unmarshal(msg, &serverAction)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot unmarshal client action: %v", err)
				return errorx.Unknown
			}

			go func() {
				engine.requestAction <- GameActionProxyRequest{
					ProxyID: proxyID,
					Actions: serverAction,
				}
			}()

		case response, ok := <-responseAction:
			if !ok {
				return errorx.New(errorx.ChangeEngine, "Engine was changed")
			}

			if err := wsClient.Write(response); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot write to ws proxy: %v", err)
				return errorx.Unknown
			}
		}
	}

	return nil
}
