package domain

import (
	"net/http"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/ws"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type WsDomain interface {
	Serve(w http.ResponseWriter, r *http.Request)
	Run()
}

type wsDomain struct {
	roomRepo repository.RoomRepository
	Hub      *ws.Hub
	cfg      config.Configs
	logger   logger.Logger
	db       *gorm.DB
	verifier *middleware.AuthVerifier
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewWsDomain(
	roomRepo repository.RoomRepository,
	cfg config.Configs,
	logger logger.Logger,
	db *gorm.DB,
	verifier *middleware.AuthVerifier,
) WsDomain {
	return &wsDomain{
		roomRepo: roomRepo,
		cfg:      cfg,
		logger:   logger,
		db:       db,
		verifier: verifier,
	}
}

func (d *wsDomain) Serve(w http.ResponseWriter, r *http.Request) {
	ctx := xcontext.NewContext(r.Context(), r, w, d.cfg, d.logger, d.db)
	if err := d.verifier.Middleware()(ctx); err != nil {
		http.Error(w, "Access token is not valid", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Unable to connect server", http.StatusInternalServerError)
		return
	}

	roomID := r.URL.Query().Get("room_id")
	userID := xcontext.GetRequestUserID(ctx)

	if userID == "" {
		http.Error(w, "User is not valid", http.StatusBadRequest)
		return
	}

	if err := d.roomRepo.GetByRoomID(ctx, roomID); err != nil {
		http.Error(w, "Room is not valid", http.StatusBadRequest)
		return
	}

	client := ws.NewClient(d.Hub, conn, roomID, &ws.Info{
		UserID: userID,
	})
	client.Register()
}

func (d *wsDomain) Run() {
	d.Hub.Run()
}
