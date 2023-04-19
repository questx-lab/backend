package domain

import (
	"context"
	"log"
	"time"

	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/ws"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type WsDomain interface {
	ServeGameClient(xcontext.Context, *model.ServeGameClientRequest) error
	Run()
	WsSubscribeHandler(context.Context, *pubsub.Pack, time.Time)
}

type wsDomain struct {
	roomRepo         repository.RoomRepository
	Hub              *ws.Hub
	verifier         *middleware.AuthVerifier
	requestPublisher pubsub.Publisher
}

func NewWsDomain(
	roomRepo repository.RoomRepository,
	verifier *middleware.AuthVerifier,
	publisher pubsub.Publisher,
) WsDomain {
	return &wsDomain{
		roomRepo:         roomRepo,
		verifier:         verifier,
		Hub:              ws.NewHub(),
		requestPublisher: publisher,
	}
}

func (d *wsDomain) ServeGameClient(ctx xcontext.Context, req *model.ServeGameClientRequest) error {
	userID := xcontext.GetRequestUserID(ctx)
	if err := d.roomRepo.GetByRoomID(ctx, req.RoomID); err != nil {
		return errorx.New(errorx.BadRequest, "Room is not valid")
	}

	client := ws.NewClient(
		d.Hub,
		ctx.GetWsConn(),
		req.RoomID,
		&ws.Info{
			UserID: userID,
		},
		func(ctx context.Context, msg []byte) {
			d.requestPublisher.Publish(ctx, "REQUEST", &pubsub.Pack{
				Key: []byte(req.RoomID),
				Msg: msg,
			})
		},
	)

	client.Register()

	return nil
}

func (d *wsDomain) WsSubscribeHandler(ctx context.Context, pack *pubsub.Pack, t time.Time) {
	roomID := string(pack.Key)
	d.Hub.BroadCastByChannel(roomID, pack.Msg)
	log.Printf("broadcast for room_id %v successful", roomID)
}

func (d *wsDomain) Run() {
	d.Hub.Run()
}
