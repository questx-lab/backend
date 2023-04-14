package domain

import (
	"log"
	"net/http"

	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/ws"

	"github.com/gorilla/websocket"
)

type WsDomain interface {
	Serve(w http.ResponseWriter, r *http.Request)
	Run()
}

type wsDomain struct {
	roomRepo repository.RoomRepository
	Hub      *ws.Hub
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewWsDomain(roomRepo repository.RoomRepository) WsDomain {
	return &wsDomain{roomRepo: roomRepo}
}

func (d *wsDomain) Serve(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	channel := r.URL.Query().Get("room_id")

	client := ws.NewClient(d.Hub, conn, channel)
	client.Register()
}

func (d *wsDomain) Run() {
	d.Hub.Run()
}
