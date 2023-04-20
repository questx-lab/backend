package game_v2

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/ws"
)

type WsHub interface {
	Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time)
	Register(client *ws.Client)
	Unregister(client *ws.Client)
}

type wsHub struct {
	hub *ws.Hub
}

func NewWsHub(hub *ws.Hub) WsHub {
	return &wsHub{
		hub: hub,
	}
}

func (h *wsHub) Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time) {
	var resp model.GameActionClientResponse
	if err := json.Unmarshal(pack.Msg, &resp); err != nil {
		log.Printf("Unable to unmarshal: %v", err)
	}
	h.hub.BroadCastByRoomID(resp.RoomID, pack.Msg)
}

func (h *wsHub) Register(client *ws.Client) {
	h.Register(client)
}
func (h *wsHub) Unregister(client *ws.Client) {
	h.Unregister(client)
}
