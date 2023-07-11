package ws

import (
	"errors"

	"github.com/gorilla/websocket"
)

type MessageInfo struct {
	msg             []byte
	needCompression bool
}

type Info struct {
	UserID    string
	Ip        string
	UserAgent string
}

type Client struct {
	Conn *websocket.Conn
	R    chan []byte
	W    chan MessageInfo
}

func NewClient(conn *websocket.Conn) *Client {
	if conn == nil {
		return nil
	}

	c := &Client{
		Conn: conn,
		R:    make(chan []byte, 128),
		W:    make(chan MessageInfo, 128),
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
			originMsg, err := Decompress(msg)
			if err != nil {
				continue
			}

			c.R <- originMsg
		}
	}
}

func (c *Client) runWriter() {
	defer close(c.W)

	for {
		msgInfo := <-c.W

		msg := msgInfo.msg
		if msgInfo.needCompression {
			var err error
			msg, err = Compress(msgInfo.msg)
			if err != nil {
				continue
			}
		}

		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

func (c *Client) Write(msg []byte, needCompression bool) (err error) {
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

	c.W <- MessageInfo{msg: msg, needCompression: needCompression}
	return nil
}
