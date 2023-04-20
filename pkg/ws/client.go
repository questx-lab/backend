package ws

import (
	"context"
	"log"

	"github.com/gorilla/websocket"
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	roomID string

	userID string

	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	handler func(context.Context, []byte)
}

func NewClient(
	conn *websocket.Conn,
	roomID string,
	userID string,
	handler func(context.Context, []byte),
) *Client {
	return &Client{
		conn:    conn,
		roomID:  roomID,
		userID:  userID,
		handler: handler,
		send:    make(chan []byte, 1<<16),
	}
}

func (c *Client) Read() {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket.IsUnexpectedCloseError: %v\n", err)
			}
			break
		}
		c.handler(context.Background(), message)
	}
}

func (c *Client) Write(msg []byte) {
	data := websocket.FormatCloseMessage(websocket.CloseNormalClosure, string(msg))
	if err := c.conn.WriteMessage(websocket.CloseMessage, data); err != nil {
		log.Printf("Unable to send message: %v\n", err)
	}
}
