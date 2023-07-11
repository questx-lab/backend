package ws

import (
	"errors"

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
	W    chan []byte
}

func NewClient(conn *websocket.Conn) *Client {
	if conn == nil {
		return nil
	}

	c := &Client{
		Conn: conn,
		R:    make(chan []byte, 128),
		W:    make(chan []byte, 128),
	}

	go c.runReader()
	go c.runWriter()
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

		if t == websocket.TextMessage {
			cmsg, err := UncompressGZIP(msg)
			if err != nil {
				continue
			}

			c.R <- cmsg
		}
	}
}
func (c *Client) runWriter() {
	defer close(c.W)

	for {
		msg := <-c.W

		cmsg, err := CompressGZIP(msg)
		if err != nil {
			continue
		}

		if err := c.Conn.WriteMessage(websocket.TextMessage, cmsg); err != nil {
			break
		}
	}
}

func (c *Client) Write(msg []byte) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		if s, ok := r.(string); ok {
			err = errors.New(s)
		} else {
			err = errors.New("connection is closed")
		}
	}()

	c.W <- msg
	return nil
}
