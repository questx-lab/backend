package ws

import (
	"github.com/gorilla/websocket"
)

type Connection struct {
	conn *websocket.Conn
	R    chan []byte
}

func NewConn(conn *websocket.Conn) *Connection {
	c := &Connection{
		conn: conn,
		R:    make(chan []byte),
	}

	go c.runReader()
	return c
}

func (c *Connection) runReader() error {
	defer close(c.R)

	for {
		messageType, p, err := c.conn.ReadMessage()
		if err != nil {
			return err
		}

		if messageType == websocket.TextMessage {
			c.R <- p
		}
	}
}

func (c *Connection) Write(msg any) error {
	switch t := msg.(type) {
	case string:
		return c.conn.WriteMessage(websocket.TextMessage, []byte(t))
	case []byte:
		return c.conn.WriteMessage(websocket.TextMessage, t)
	default:
		return c.conn.WriteJSON(t)
	}
}
