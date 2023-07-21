package ws

import (
	"github.com/gorilla/websocket"
)

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
		R:    make(chan []byte, 16),
	}

	go c.runReader()
	return c
}

func (c *Client) runReader() {
	defer close(c.R)

	for {
		t, msg, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}

		if t == websocket.CloseMessage {
			return
		}

		if t == websocket.TextMessage || t == websocket.BinaryMessage {
			originMsg, err := Decompress(msg)
			if err != nil {
				continue
			}

			c.R <- originMsg
		}
	}
}

func (c *Client) Write(msg []byte, needCompression bool) error {
	if needCompression {
		var err error
		msg, err = Compress(msg)
		if err != nil {
			return nil
		}
	}

	if err := c.Conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
		return err
	}

	return nil
}
