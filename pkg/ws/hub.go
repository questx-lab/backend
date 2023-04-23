package ws

import (
	"github.com/puzpuzpuz/xsync"
)

type Hub struct {
	clients       *xsync.MapOf[string, *ClientV2]
	roomClientV2s map[string]*xsync.MapOf[string, *ClientV2]
	register      chan *ClientV2
	unregister    chan *ClientV2
}

func NewHub() *Hub {
	return &Hub{
		register:      make(chan *ClientV2),
		unregister:    make(chan *ClientV2),
		clients:       xsync.NewMapOf[*ClientV2](),
		roomClientV2s: make(map[string]*xsync.MapOf[string, *ClientV2]),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients.Store(client.userID, client)
			if _, ok := h.roomClientV2s[client.roomID]; !ok {
				h.roomClientV2s[client.roomID] = xsync.NewMapOf[*ClientV2]()
			}
			h.roomClientV2s[client.roomID].Store(client.userID, client)
		case client := <-h.unregister:
			h.disconnect(client)
		}
	}
}

func (h *Hub) disconnect(client *ClientV2) {
	h.clients.LoadAndDelete(client.userID)
	h.roomClientV2s[client.roomID].LoadAndDelete(client.userID)
	close(client.send)
}

func (h *Hub) BroadCastByRoomID(roomID string, message []byte) {
	h.roomClientV2s[roomID].Range(func(userID string, client *ClientV2) bool {
		client.Write(message)
		return true
	})
}

func (h *Hub) Register(client *ClientV2) {
	h.register <- client
}

func (h *Hub) Unregister(client *ClientV2) {
	h.unregister <- client
}
