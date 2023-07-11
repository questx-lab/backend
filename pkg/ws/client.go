package ws

import (
	"encoding/base64"
	"errors"
	"log"
	"time"

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

	i := 0
	for {
		i++
		t, msg, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}

		if t == websocket.CloseMessage {
			return
		}

		if t == websocket.TextMessage {
			decodedMsg := make([]byte, base64.StdEncoding.DecodedLen(len(msg)))
			if _, err := base64.StdEncoding.Decode(decodedMsg, msg); err != nil {
				continue
			}

			start := time.Now()
			originMsg, err := DecompressFlate(decodedMsg)
			if err != nil {
				continue
			}
			if len(msg) > 1024 && i%100 == 0 {
				log.Println("DECOMPRESS CHECK", time.Since(start), len(msg), len(c.R))
			}

			c.R <- originMsg
		}
	}
}

func (c *Client) runWriter() {
	defer close(c.W)

	i := 0
	for {
		i++
		msg := <-c.W

		start := time.Now()
		cmsg, err := CompressFlate(msg)
		if err != nil {
			continue
		}
		if len(msg) > 1024 && i%100 == 0 {
			log.Println("COMPRESS CHECK", time.Since(start), len(msg), len(c.W))
		}

		encodedMsg := make([]byte, base64.StdEncoding.EncodedLen(len(cmsg)))
		base64.StdEncoding.Encode(encodedMsg, cmsg)

		if err := c.Conn.WriteMessage(websocket.TextMessage, encodedMsg); err != nil {
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
