package ws

import (
	"context"
	"log"

	"github.com/gorilla/websocket"
)

// Client is a middleman between the websocket connection and the hub.
type ClientV2 struct {
	roomID string

	userID string

	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	handler func(context.Context, []byte)
}

func NewClientV2(
	conn *websocket.Conn,
	roomID string,
	userID string,
	handler func(context.Context, []byte),
) *ClientV2 {
	return &ClientV2{
		conn:    conn,
		roomID:  roomID,
		userID:  userID,
		handler: handler,
		send:    make(chan []byte, 1<<16),
	}
}

func (c *ClientV2) Read() {
	for {
		_, message, err := c.conn.ReadMessage()
		log.Println(string(message))
		if err != nil {
			log.Println(err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket.IsUnexpectedCloseError: %v\n", err)
			}
			continue
		}
		c.handler(context.Background(), message)
	}
}

func (c *ClientV2) Write(msg []byte) {
	if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Printf("Unable to send message: %v\n", err)
	}
}
