package gameengine

import (
	"context"
	"encoding/json"
	"time"

	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type engine struct {
	logger logger.Logger

	gamestate     *GameState
	pendingAction <-chan model.GameActionServerRequest
	publisher     pubsub.Publisher
	gameRepo      repository.GameRepository
}

func NewEngine(
	ctx xcontext.Context,
	engineRouter Router,
	publisher pubsub.Publisher,
	logger logger.Logger,
	gameRepo repository.GameRepository,
	roomID string,
) (*engine, error) {
	gamestate, err := newGameState(ctx, gameRepo, roomID)
	if err != nil {
		return nil, err
	}

	err = gamestate.LoadUser(ctx, gameRepo)
	if err != nil {
		return nil, err
	}

	pendingAction, err := engineRouter.Register(roomID)
	if err != nil {
		return nil, err
	}

	engine := &engine{
		logger:        logger,
		gamestate:     gamestate,
		pendingAction: pendingAction,
		publisher:     publisher,
		gameRepo:      gameRepo,
	}

	go engine.runUpdateDatabase(ctx)
	go engine.run()

	return engine, nil
}

func (e *engine) run() {
	e.logger.Infof("Game engine for room %s is started", e.gamestate.roomID)

	for {
		actionRequest, ok := <-e.pendingAction
		if !ok {
			break
		}

		action, err := parseAction(actionRequest)
		if err != nil {
			e.logger.Debugf("Cannot parse action: %v", err)
			continue
		}

		err = e.gamestate.Apply(action)
		if err != nil {
			e.logger.Debugf("Cannot apply action to room %s: %v", e.gamestate.roomID, err)
			continue
		}

		actionResponse, err := formatAction(action)
		if err != nil {
			e.logger.Errorf("Cannot format action response: %v", err)
			continue
		}

		b, err := json.Marshal(actionResponse)
		if err != nil {
			e.logger.Errorf("Cannot marshal action response: %v", err)
			continue
		}

		err = e.publisher.Publish(context.Background(), model.ResponseTopic, &pubsub.Pack{
			Key: []byte(e.gamestate.roomID),
			Msg: b,
		})
		if err != nil {
			e.logger.Errorf("Cannot publish action response: %v", err)
			continue
		}
	}

	e.logger.Infof("Game engine for room %s is stopped", e.gamestate.roomID)
}

func (e *engine) runUpdateDatabase(ctx xcontext.Context) {
	for range time.Tick(ctx.Configs().Game.UpdateDatabaseEvery) {
		e.updateDatabase(ctx)
	}
}

func (e *engine) updateDatabase(ctx xcontext.Context) {
	users := e.gamestate.UserDiff()
	if len(users) > 0 {
		for _, user := range users {
			err := e.gameRepo.UpsertGameUser(ctx, user)
			if err != nil {
				ctx.Logger().Errorf("Cannot upsert game user: %v", err)
			}
		}
		ctx.Logger().Infof("Update database for game state successfully")
	}
}
