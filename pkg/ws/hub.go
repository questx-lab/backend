package ws

// Hub maintains the set of active clients and broadcasts messages to the
// clients.

type clients map[*Client]bool

type Hub struct {
	// Registered clients.
	clients clients

	channels map[string]clients

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		channels:   make(map[string]clients),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			if _, ok := h.channels[client.channel]; !ok {
				h.channels[client.channel] = make(clients)
			}
			h.channels[client.channel][client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.disconnect(client)
			}
		}
	}
}

func (h *Hub) disconnect(client *Client) {
	delete(h.clients, client)
	delete(h.channels[client.channel], client)
	close(client.send)
}

func (h *Hub) BroadCastByChannel(channel string, message []byte) {
	for client := range h.channels[channel] {
		select {
		case client.send <- message:
		default:
			h.disconnect(client)
		}
	}
}
