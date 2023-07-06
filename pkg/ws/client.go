package ws

import (
	"github.com/gorilla/websocket"
)

type Info struct {
	UserID    string
	Ip        string
	UserAgent string
}

type Client struct {
	Conn *websocket.Conn
	R    chan []byte
}

func NewClient(conn *websocket.Conn) *Client {
	if conn == nil {
		return nil
	}

	c := &Client{
		Conn: conn,
		R:    make(chan []byte),
	}

	go c.runReader()
	return c
}

func (c *Client) runReader() {
	defer close(c.R)

	for {
		messageType, p, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}

		if messageType == websocket.CloseMessage {
			return
		}

		if messageType == websocket.TextMessage {
			c.R <- p
		}
	}
}

func (c *Client) Write(msg any) error {
	switch t := msg.(type) {
	case string:
		return c.Conn.WriteMessage(websocket.TextMessage, []byte(t))
	case []byte:
		return c.Conn.WriteMessage(websocket.TextMessage, t)
	default:
		return c.Conn.WriteJSON(t)
	}
}
