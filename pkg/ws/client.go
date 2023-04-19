package ws

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Info struct {
	UserID    string
	Ip        string
	UserAgent string
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	info *Info

	hub *Hub

	channel string

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// handle if message sent
	handler func(context.Context, []byte)
}

func NewClient(
	hub *Hub,
	conn *websocket.Conn,
	channel string,
	info *Info,
	handler func(context.Context, []byte),
) *Client {
	return &Client{
		hub:     hub,
		conn:    conn,
		channel: channel,
		send:    make(chan []byte, 256),
		info:    info,
		handler: handler,
	}
}

func (c *Client) Register() {
	c.hub.register <- c
	go c.writePump()
	go c.readPump()
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Printf("c.conn.SetReadDeadline: %v\n", err)
	}
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			log.Printf("c.conn.SetReadDeadline: %v\n", err)
			return err
		}
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket.IsUnexpectedCloseError: %v\n", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// handle data if client send
		c.handler(context.Background(), message)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.conn.Close(); err != nil {
			log.Printf("c.conn.Close: %v\n", err)
		}
	}()
	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("c.conn.SetWriteDeadline: %v\n", err)
			}
			if !ok {
				// The hub closed the channel.
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Printf("c.conn.WriteMessage: %v\n", err)
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("c.conn.NextWriter: %v\n", err)
				return
			}
			if _, err := w.Write(message); err != nil {
				log.Printf("w.Write: %v\n", err)
			}

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				if _, err := w.Write(newline); err != nil {
					log.Printf("w.Write, newline: %v\n", err)
				}
				if _, err := w.Write(<-c.send); err != nil {
					log.Printf("w.Write, c.send: %v\n", err)
				}
			}

			if err := w.Close(); err != nil {
				log.Printf("w.Close: %v\n", err)
				return
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("c.conn.SetWriteDeadline: %v\n", err)
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("c.conn.WriteMessage: %v\n", err)
				return
			}
		}
	}
}
