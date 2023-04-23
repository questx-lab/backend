package domain

import (
	"context"
	"encoding/json"
	"log"

	"github.com/questx-lab/backend/internal/domain/gameproxy"
	"github.com/questx-lab/backend/internal/domain/gamestate"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/ws"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type GameProxyDomain interface {
	ServeGameClient(xcontext.Context, *model.ServeGameClientRequest) error
	ServeGameClientV2(xcontext.Context, *model.ServeGameClientRequest) error
}

type gameProxyDomain struct {
	gameRepo   repository.GameRepository
	publisher  pubsub.Publisher
	hub        *ws.Hub
	gameRouter gameproxy.GameRouter
	gameHubs   map[string]gameproxy.GameHub
}

func NewGameProxyDomain(
	gameRepo repository.GameRepository,
	publisher pubsub.Publisher,
	hub *ws.Hub,
) GameProxyDomain {
	gameRouter := gameproxy.NewGameRouter()
	go gameRouter.Run()
	return &gameProxyDomain{
		gameRepo:   gameRepo,
		publisher:  publisher,
		gameRouter: gameRouter,
		gameHubs:   make(map[string]gameproxy.GameHub),
		hub:        hub,
	}
}

func (d *gameProxyDomain) ServeGameClient(ctx xcontext.Context, req *model.ServeGameClientRequest) error {
	userID := xcontext.GetRequestUserID(ctx)
	room, err := d.gameRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return errorx.New(errorx.BadRequest, "Room is not valid")
	}
	gameMap, err := d.gameRepo.GetMapByID(ctx, room.MapID)
	if err != nil {
		return errorx.Unknown
	}

	mapContent := model.GameActionClientResponse{
		Type: "map",
		Value: map[string]any{
			"content": string(gameMap.Content),
		},
	}

	err = ctx.WsClient().Write(mapContent)
	if err != nil {
		ctx.Logger().Errorf("Cannot write to ws: %v", err)
		return errorx.Unknown
	}

	if _, ok := d.gameHubs[req.RoomID]; !ok {
		hub, err := gameproxy.NewGameHub(ctx, d.gameRepo, req.RoomID)
		if err != nil {
			ctx.Logger().Errorf("Cannot create game hub: %v", err)
			return errorx.Unknown
		}

		d.gameHubs[req.RoomID] = hub
		go hub.Run()

		aggregator, err := gameproxy.NewGameAggregator(req.RoomID, d.gameRouter, hub)
		if err != nil {
			ctx.Logger().Errorf("Cannot create game aggregator: %v", err)
			return errorx.Unknown
		}

		go aggregator.Run()
	}

	// Register to hub to receive broadcast messages.
	hub := d.gameHubs[req.RoomID]
	hubChannel, err := hub.Register(userID)
	if err != nil {
		ctx.Logger().Debugf("Cannot register user to hub: %v", err)
		return errorx.Unknown
	}

	// Join the user in room.
	err = d.gameRouter.Route(model.GameActionRouterRequest{
		RoomID: req.RoomID,
		UserID: userID,
		Type:   gamestate.JoinActionType,
	})

	if err != nil {
		ctx.Logger().Errorf("Cannot join user in room: %v", err)
		return errorx.Unknown
	}

	defer func() {
		// Remove user from room.
		err = d.gameRouter.Route(model.GameActionRouterRequest{
			RoomID: req.RoomID,
			UserID: userID,
			Type:   gamestate.ExitActionType,
		})

		if err != nil {
			ctx.Logger().Errorf("Cannot exit user from room: %v", err)
		}

		// Unregister this client from hub.
		hub.Unregister(userID)
	}()

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
			// request publisher
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

func (d *gameProxyDomain) ServeGameClientV2(ctx xcontext.Context, req *model.ServeGameClientRequest) error {
	userID := xcontext.GetRequestUserID(ctx)
	room, err := d.gameRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return errorx.New(errorx.BadRequest, "Room is not valid")
	}

	gameMap, err := d.gameRepo.GetMapByID(ctx, room.MapID)
	if err != nil {
		return errorx.Unknown
	}
	client := ws.NewClientV2(ctx.WsConn(), req.RoomID, userID, func(ctx context.Context, msg []byte) {
		var clientReq model.GameActionClientRequest
		if err := json.Unmarshal(msg, &clientReq); err != nil {
			log.Printf("Unable to unmarshal client request: %v\n", err)
			return
		}

		serverReq := model.GameActionServerRequest{
			Type:   clientReq.Type,
			Value:  clientReq.Value,
			UserID: userID,
		}

		b, err := json.Marshal(&serverReq)
		if err != nil {
			log.Printf("Unable to marshal server request: %v\n", err)
			return
		}
		if err := d.publisher.Publish(ctx, model.RequestTopic, &pubsub.Pack{
			Key: []byte(req.RoomID),
			Msg: b,
		}); err != nil {
			log.Printf("Unable to publish: %v\n", err)
		}
	})

	client.Write(gameMap.Content)

	d.hub.Register(client)
	defer d.hub.Unregister(client)
	client.Read()
	return nil
}
