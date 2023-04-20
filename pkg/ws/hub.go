package ws

import (
	"github.com/puzpuzpuz/xsync"
)

type Hub struct {
	clients     *xsync.MapOf[string, *Client]
	roomClients map[string]*xsync.MapOf[string, *Client]
	register    chan *Client
	unregister  chan *Client
}

func NewHub() *Hub {
	return &Hub{
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     xsync.NewMapOf[*Client](),
		roomClients: make(map[string]*xsync.MapOf[string, *Client]),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients.Store(client.userID, client)
			if _, ok := h.roomClients[client.roomID]; !ok {
				h.roomClients[client.roomID] = xsync.NewMapOf[*Client]()
			}
			h.roomClients[client.roomID].Store(client.userID, client)
		case client := <-h.unregister:
			h.disconnect(client)
		}
	}
}

func (h *Hub) disconnect(client *Client) {
	h.clients.LoadAndDelete(client.userID)
	h.roomClients[client.roomID].LoadAndDelete(client.userID)
	close(client.send)
}

func (h *Hub) BroadCastByRoomID(roomID string, message []byte) {
	h.roomClients[roomID].Range(func(userID string, client *Client) bool {
		client.Write(message)
		return true
	})
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}
