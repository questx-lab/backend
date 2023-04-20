package domain

import (
	"context"
	"log"

	"github.com/questx-lab/backend/internal/domain/game_v2"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/ws"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type WsDomain interface {
	// ServeGameClient(xcontext.Context, *model.ServeGameClientRequest) error
	ServeGameClientV2(xcontext.Context, *model.ServeGameClientRequest) error
}

type wsDomain struct {
	gameRepo  repository.GameRepository
	hub       game_v2.WsHub
	publisher pubsub.Publisher
}

func NewWsDomain(
	gameRepo repository.GameRepository,
	hub game_v2.WsHub,
	publisher pubsub.Publisher,
) WsDomain {
	return &wsDomain{
		gameRepo:  gameRepo,
		hub:       hub,
		publisher: publisher,
	}
}

func (d *wsDomain) ServeGameClientV2(ctx xcontext.Context, req *model.ServeGameClientRequest) error {
	userID := xcontext.GetRequestUserID(ctx)
	room, err := d.gameRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return errorx.New(errorx.BadRequest, "Room is not valid")
	}

	gameMap, err := d.gameRepo.GetMapByID(ctx, room.MapID)
	if err != nil {
		return errorx.Unknown
	}

	client := ws.NewClient(ctx.WsConn(), req.RoomID, userID, func(ctx context.Context, msg []byte) {
		if err := d.publisher.Publish(ctx, string(model.RequestTopic), &pubsub.Pack{
			Key: []byte(req.RoomID),
			Msg: msg,
		}); err != nil {
			log.Printf("Unable to publish: %v\n", err)
		}
	})

	client.Write(gameMap.Content)

	d.hub.Register(client)
	defer d.hub.Unregister(client)

	go client.Read()
	return nil
}
