package gameengine

import (
	"context"
	"encoding/json"
	"time"

	"github.com/questx-lab/backend/internal/domain/statistic"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/buffer"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type engine struct {
	gamestate     *GameState
	pendingAction <-chan []model.GameActionServerRequest
	publisher     pubsub.Publisher
	gameRepo      repository.GameRepository
}

func NewEngine(
	ctx context.Context,
	engineRouter Router,
	publisher pubsub.Publisher,
	gameRepo repository.GameRepository,
	userRepo repository.UserRepository,
	followerRepo repository.FollowerRepository,
	leaderboard statistic.Leaderboard,
	storage storage.Storage,
	roomID string,
) (*engine, error) {
	gamestate, err := newGameState(ctx, gameRepo, userRepo, followerRepo, leaderboard, storage, roomID)
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

	pendingAction, err := engineRouter.Register(ctx, roomID)
	if err != nil {
		return nil, err
	}

	engine := &engine{
		gamestate:     gamestate,
		pendingAction: pendingAction,
		publisher:     publisher,
		gameRepo:      gameRepo,
	}

	go engine.runUpdateDatabase(ctx)
	go engine.run(ctx)

	return engine, nil
}

func (e *engine) run(ctx context.Context) {
	xcontext.Logger(ctx).Infof("Game engine for room %s is started", e.gamestate.roomID)

	ticker := time.NewTicker(xcontext.Configs(ctx).Game.EngineBatchingFrequency)
	pendingMsg := [][]byte{}
	isStop := false
	for !isStop {
		select {
		case actionRequests, ok := <-e.pendingAction:
			if !ok {
				isStop = true
				break
			}

			for _, req := range actionRequests {
				action, err := parseAction(req)
				if err != nil {
					xcontext.Logger(ctx).Debugf("Cannot parse action: %v", err)
					continue
				}

				err = e.gamestate.Apply(ctx, action)
				if err != nil {
					xcontext.Logger(ctx).Debugf("Cannot apply action to room %s: %v", e.gamestate.roomID, err)
					continue
				}

				actionResponse, err := formatAction(action)
				if err != nil {
					xcontext.Logger(ctx).Errorf("Cannot format action response: %v", err)
					continue
				}

				b, err := json.Marshal(actionResponse)
				if err != nil {
					xcontext.Logger(ctx).Errorf("Cannot marshal action response: %v", err)
					continue
				}

				pendingMsg = append(pendingMsg, b)
			}

		case <-ticker.C:
			if len(pendingMsg) == 0 {
				continue
			}

			buf := buffer.New()
			buf.AppendByte('[')

			for i, msg := range pendingMsg {
				buf.AppendBytes(msg)
				if i < len(pendingMsg)-1 {
					buf.AppendByte(',')
				}
			}

			pendingMsg = pendingMsg[:0]
			buf.AppendByte(']')
			err := e.publisher.Publish(ctx, model.GameActionResponseTopic, &pubsub.Pack{
				Key: []byte(e.gamestate.roomID),
				Msg: buf.Bytes(),
			})
			buf.Free()
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot publish action response: %v", err)
				continue
			}
		}
	}

	xcontext.Logger(ctx).Infof("Game engine for room %s is stopped", e.gamestate.roomID)
}

func (e *engine) runUpdateDatabase(ctx context.Context) {
	for range time.Tick(xcontext.Configs(ctx).Game.GameSaveFrequency) {
		e.updateDatabase(ctx)
	}
}

func (e *engine) updateDatabase(ctx context.Context) {
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
			err := e.gameRepo.UpsertLuckybox(ctx, luckybox)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot upsert luckybox: %v", err)
			}
		}

		xcontext.Logger(ctx).Infof("Update database for luckybox successfully")
	}
}
