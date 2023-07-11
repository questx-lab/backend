package gameengine

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/statistic"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/ws"
	"github.com/questx-lab/backend/pkg/xcontext"
)

const maxActionChannelSize = 1 << 13

type GameActionProxyRequest struct {
	ProxyID string
	Actions []model.GameActionServerRequest
}

type engine struct {
	done             chan any
	gamestate        *GameState
	requestAction    chan GameActionProxyRequest
	responseMsg      chan []byte
	proxyChannels    map[string]chan []byte
	proxyMutex       sync.Mutex
	gameRepo         repository.GameRepository
	gameLuckyboxRepo repository.GameLuckyboxRepository
}

func NewEngine(
	ctx context.Context,
	gameRepo repository.GameRepository,
	gameLuckyboxRepo repository.GameLuckyboxRepository,
	gameCharacterRepo repository.GameCharacterRepository,
	userRepo repository.UserRepository,
	followerRepo repository.FollowerRepository,
	leaderboard statistic.Leaderboard,
	storage storage.Storage,
	roomID string,
) (*engine, error) {
	gamestate, err := newGameState(ctx, gameRepo, gameLuckyboxRepo, gameCharacterRepo,
		userRepo, followerRepo, leaderboard, storage, roomID)
	if err != nil {
		return nil, err
	}

	err = gamestate.LoadUser(ctx)
	if err != nil {
		return nil, err
	}

	err = gamestate.LoadLuckybox(ctx)
	if err != nil {
		return nil, err
	}

	engine := &engine{
		gamestate:        gamestate,
		done:             make(chan any),
		requestAction:    make(chan GameActionProxyRequest, maxActionChannelSize),
		responseMsg:      make(chan []byte, maxActionChannelSize),
		proxyChannels:    make(map[string]chan []byte),
		proxyMutex:       sync.Mutex{},
		gameRepo:         gameRepo,
		gameLuckyboxRepo: gameLuckyboxRepo,
	}

	go engine.run(ctx)
	go engine.serveProxy(ctx)
	go engine.updateDatabase(ctx)

	time.AfterFunc(20*time.Second, func() {
		engine.proxyMutex.Lock()
		defer engine.proxyMutex.Unlock()

		xcontext.Logger(ctx).Infof("Cleanup old disconnected proxy")
		engine.requestAction <- GameActionProxyRequest{
			Actions: []model.GameActionServerRequest{{
				UserID: "",
				Type:   CleanupProxyAction{}.Type(),
				Value:  map[string]any{"live_proxy_ids": common.MapKeys(engine.proxyChannels)},
			}},
		}
	})

	return engine, nil
}

func (e *engine) Stop(ctx context.Context) {
	e.proxyMutex.Lock()
	defer e.proxyMutex.Unlock()

	close(e.done)
	for i := range e.proxyChannels {
		close(e.proxyChannels[i])
	}
	e.proxyChannels = nil
}

func (e *engine) RegisterProxy(ctx context.Context, id string) <-chan []byte {
	e.proxyMutex.Lock()
	defer e.proxyMutex.Unlock()

	if _, ok := e.proxyChannels[id]; ok {
		xcontext.Logger(ctx).Errorf("Proxy hub %s is already existed in engine %s", id, e.gamestate.roomID)
		return nil
	}

	c := make(chan []byte, maxActionChannelSize)
	e.proxyChannels[id] = c
	xcontext.Logger(ctx).Infof("Proxy hub %s registered to %s successfully", id, e.gamestate.roomID)
	return c
}

func (e *engine) UnregisterProxy(ctx context.Context, id string) {
	e.proxyMutex.Lock()
	defer e.proxyMutex.Unlock()

	c, ok := e.proxyChannels[id]
	if !ok {
		xcontext.Logger(ctx).Errorf("Not found proxy hub %s in engine %s", id, e.gamestate.roomID)
		return
	}

	close(c)
	delete(e.proxyChannels, id)

	go func() {
		e.requestAction <- GameActionProxyRequest{
			Actions: []model.GameActionServerRequest{{
				UserID: "",
				Type:   CleanupProxyAction{}.Type(),
				Value:  map[string]any{"live_proxy_ids": common.MapKeys(e.proxyChannels)},
			}},
		}
	}()

	xcontext.Logger(ctx).Infof("Proxy hub %s unregistered from %s", id, e.gamestate.roomID)
}

func (e *engine) run(ctx context.Context) {
	xcontext.Logger(ctx).Infof("Game engine for room %s is started", e.gamestate.roomID)
	ticker := time.NewTicker(time.Millisecond)
	defer func() {
		close(e.responseMsg)
		ticker.Stop()
	}()

	pendingResponse := []model.GameActionServerResponse{}
	isStop := false
	for !isStop {
		select {
		case actionRequests := <-e.requestAction:
			for _, req := range actionRequests.Actions {
				action, err := parseAction(req)
				if err != nil {
					xcontext.Logger(ctx).Debugf("Cannot parse action: %v", err)
					continue
				}

				replyActions, err := e.gamestate.Apply(ctx, actionRequests.ProxyID, action)
				if err != nil {
					xcontext.Logger(ctx).Debugf("Cannot apply action %s to room %s: %v",
						action.Type(), e.gamestate.roomID, err)
					continue
				}

				actionResponse, err := formatAction(action)
				if err != nil {
					xcontext.Logger(ctx).Errorf("Cannot format action response: %v", err)
					continue
				}

				pendingResponse = append(pendingResponse, actionResponse)

				if len(replyActions) > 0 {
					go func() {
						e.requestAction <- GameActionProxyRequest{
							ProxyID: actionRequests.ProxyID,
							Actions: replyActions,
						}
					}()
				}
			}

		case <-ticker.C:
			if len(pendingResponse) == 0 {
				continue
			}

			batch := common.Batch(&pendingResponse, 1024)
			common.DetectBottleneck(ctx, batch, pendingResponse, "sending responses to proxy")
			b, err := json.Marshal(batch)
			if err != nil {
				xcontext.Logger(ctx).Warnf("Cannot marshal response: %v", err)
				continue
			}

			compressedBatch, err := ws.Compress(b)
			if err != nil {
				xcontext.Logger(ctx).Warnf("Cannot compress response: %v", err)
				continue
			}

			e.responseMsg <- compressedBatch

		case <-e.done:
			isStop = true
		}
	}

	xcontext.Logger(ctx).Infof("Game engine for room %s is stopped", e.gamestate.roomID)
}

func (e *engine) serveProxy(ctx context.Context) {
	for {
		b, ok := <-e.responseMsg
		if !ok {
			break
		}

		e.proxyMutex.Lock()
		for i := range e.proxyChannels {
			e.proxyChannels[i] <- b
		}
		e.proxyMutex.Unlock()
	}
}

func (e *engine) updateDatabase(ctx context.Context) {
	select {
	case <-e.done:
		// No need to update database anymore.

	default:
		defer time.AfterFunc(xcontext.Configs(ctx).Game.GameSaveFrequency, func() {
			e.updateDatabase(ctx)
		})

		users := e.gamestate.UserDiff()
		if len(users) > 0 {
			for _, user := range users {
				err := e.gameRepo.UpsertGameUser(ctx, user)
				if err != nil {
					xcontext.Logger(ctx).Errorf("Cannot upsert game user: %v", err)
				}
			}
			xcontext.Logger(ctx).Infof("Update database for game user successfully")
		}

		luckyboxes := e.gamestate.LuckyboxDiff()
		if len(luckyboxes) > 0 {
			for _, luckybox := range luckyboxes {
				err := e.gameLuckyboxRepo.UpsertLuckybox(ctx, luckybox)
				if err != nil {
					xcontext.Logger(ctx).Errorf("Cannot upsert luckybox: %v", err)
				}
			}

			xcontext.Logger(ctx).Infof("Update database for luckybox successfully")
		}
	}
}
