package domain

import (
	"encoding/json"

	"github.com/questx-lab/backend/internal/domain/game"
	"github.com/questx-lab/backend/internal/domain/gamestate"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type WsDomain interface {
	ServeGameClient(xcontext.Context, *model.ServeGameClientRequest) error
}

type wsDomain struct {
	gameRepo   repository.GameRepository
	gameRouter game.GameRouter
	gameHubs   map[string]game.GameHub
}

func NewWsDomain(
	gameRepo repository.GameRepository,
) WsDomain {
	gameRouter := game.NewGameRouter()
	go gameRouter.Run()
	return &wsDomain{
		gameRepo:   gameRepo,
		gameRouter: gameRouter,
		gameHubs:   make(map[string]game.GameHub),
	}
}

func (d *wsDomain) ServeGameClient(ctx xcontext.Context, req *model.ServeGameClientRequest) error {
	userID := xcontext.GetRequestUserID(ctx)
	room, err := d.gameRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return errorx.New(errorx.BadRequest, "Room is not valid")
	}

	gameMap, err := d.gameRepo.GetMapByID(ctx, room.MapID)
	if err != nil {
		return errorx.Unknown
	}

	if _, ok := d.gameHubs[req.RoomID]; !ok {
		hub, err := game.NewGameHub(ctx, d.gameRepo, req.RoomID)
		if err != nil {
			ctx.Logger().Errorf("Cannot create game hub: %v", err)
			return errorx.Unknown
		}

		d.gameHubs[req.RoomID] = hub
		go hub.Run()

		aggregator, err := game.NewGameAggregator(req.RoomID, d.gameRouter, hub)
		if err != nil {
			ctx.Logger().Errorf("Cannot create game aggregator: %v", err)
			return errorx.Unknown
		}

		go aggregator.Run()
	}

	hub := d.gameHubs[req.RoomID]
	hubChannel, err := hub.Register(userID)
	if err != nil {
		ctx.Logger().Debugf("Cannot register user to hub: %v", err)
		return errorx.Unknown
	}

	defer hub.Unregister(userID)

	err = ctx.WsClient().Write(gameMap.Content)
	if err != nil {
		ctx.Logger().Errorf("Cannot write to ws: %v", err)
		return errorx.Unknown
	}

	isStop := false
	for !isStop {
		select {
		case msg, ok := <-ctx.WsClient().R:
			if !ok {
				isStop = true
				break
			}

			action := model.GameActionClientRequest{}
			err := json.Unmarshal(msg, &action)
			if err != nil {
				ctx.Logger().Errorf("Cannot unmarshal action: %v", err)
				return errorx.Unknown
			}

			routerAction := gamestate.ClientActionToRouterAction(action, req.RoomID, userID)
			err = d.gameRouter.Route(routerAction)
			if err != nil {
				ctx.Logger().Debugf("Cannot route action: %v", err)
				return errorx.Unknown
			}

		case msg := <-hubChannel:
			err := ctx.WsClient().Write(msg)
			if err != nil {
				ctx.Logger().Errorf("Cannot write to ws: %v", err)
				return errorx.Unknown
			}
		}
	}

	return nil
}
