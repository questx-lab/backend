package domain

import (
	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/ws"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/gorilla/websocket"
)

type WsDomain interface {
	Serve(xcontext.Context) error
	Run()
}

type wsDomain struct {
	roomRepo repository.RoomRepository
	Hub      *ws.Hub
	verifier *middleware.AuthVerifier
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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

func (d *wsDomain) Serve(ctx xcontext.Context) error {
	w := ctx.Writer()
	r := ctx.Request()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return errorx.New(errorx.BadRequest, "Unable to connect server")
	}
	userID, err := d.verifyUser(ctx)
	if err != nil {
		return err
	}

	roomID := ctx.Request().URL.Query().Get("room_id")
	if err := d.roomRepo.GetByRoomID(ctx, roomID); err != nil {
		return errorx.New(errorx.BadRequest, "Room is not valid")
	}

	client := ws.NewClient(d.Hub, conn, roomID, &ws.Info{
		UserID: userID,
	})

	client.Register()
	return nil
}

func (d *wsDomain) Run() {
	d.Hub.Run()
}

func (d *wsDomain) verifyUser(ctx xcontext.Context) (string, error) {
	if err := d.verifier.Middleware()(ctx); err != nil {
		return "", errorx.New(errorx.BadRequest, "Access token is not valid")
	}

	userID := xcontext.GetRequestUserID(ctx)

	if userID == "" {
		return "", errorx.New(errorx.BadRequest, "User is not valid")
	}

	return userID, nil
}
