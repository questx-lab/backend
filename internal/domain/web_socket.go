package domain

import (
	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/ws"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type WsDomain interface {
	ServeGameClient(xcontext.Context, *model.ServeGameClientRequest) error
	Run()
}

type wsDomain struct {
	roomRepo repository.RoomRepository
	Hub      *ws.Hub
	verifier *middleware.AuthVerifier
}

func NewWsDomain(
	roomRepo repository.RoomRepository,
	verifier *middleware.AuthVerifier,
) WsDomain {
	return &wsDomain{
		roomRepo: roomRepo,
		verifier: verifier,
		Hub:      ws.NewHub(),
	}
}

func (d *wsDomain) ServeGameClient(ctx xcontext.Context, req *model.ServeGameClientRequest) error {
	userID := xcontext.GetRequestUserID(ctx)

	if err := d.roomRepo.GetByRoomID(ctx, req.RoomID); err != nil {
		return errorx.New(errorx.BadRequest, "Room is not valid")
	}

	client := ws.NewClient(d.Hub, ctx.GetWsConn(), req.RoomID, &ws.Info{
		UserID: userID,
	})

	client.Register()
	return nil
}

func (d *wsDomain) Run() {
	d.Hub.Run()
}
