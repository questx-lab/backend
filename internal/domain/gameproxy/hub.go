package gameproxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/ws"
	"github.com/questx-lab/backend/pkg/xcontext"
)

const maxMsgSize = 1 << 8

type Hub interface {
	Register(ctx context.Context, clientID string) (<-chan []byte, error)
	Unregister(ctx context.Context, clientID string) error
	ForwardSingleAction(ctx context.Context, action model.GameActionServerRequest)
}

type hub struct {
	roomID string

	pendingResponseMsg chan []model.GameActionServerResponse
	pendingRequestMsg  chan model.GameActionServerRequest

	// These following attributes need to be protected by mutex lock.
	mutex     sync.Mutex
	isRunning bool
	wsClient  *ws.Client
	gameRepo  repository.GameRepository
	clients   map[string]chan<- []byte
}

func NewHub(
	ctx context.Context,
	gameRepo repository.GameRepository,
	roomID string,
) *hub {
	hub := &hub{
		roomID:             roomID,
		gameRepo:           gameRepo,
		pendingResponseMsg: make(chan []model.GameActionServerResponse, maxMsgSize),
		pendingRequestMsg:  make(chan model.GameActionServerRequest, maxMsgSize<<4),

		mutex:     sync.Mutex{},
		isRunning: false,
		wsClient:  nil,
		clients:   make(map[string]chan<- []byte),
	}

	return hub
}

// Register allows a client to subcribe to hub. All broadcasting actions will be
// sent to this client after this point of time.
func (h *hub) Register(ctx context.Context, clientID string) (<-chan []byte, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if !h.isRunning {
		h.isRunning = true
		go h.run(ctx)
		go h.broadcast(ctx)
		go h.runForward(ctx)
	}

	var c chan []byte
	if _, ok := h.clients[clientID]; !ok {
		c = make(chan []byte, maxMsgSize)
		h.clients[clientID] = c
	} else {
		return nil, errors.New("the game client has already registered")
	}

	xcontext.Logger(ctx).Infof("User %s registered to hub %s successfully (%d)",
		clientID, h.roomID, len(h.clients))

	return c, nil
}

// Unregister removes the game client from this hub.
func (h *hub) Unregister(ctx context.Context, clientID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	c, ok := h.clients[clientID]
	if !ok {
		return errors.New("the client has not registered yet")
	}

	close(c)
	delete(h.clients, clientID)

	// Temporarily unregister hub from router.
	if h.isRunning && len(h.clients) == 0 {
		h.isRunning = false
		h.wsClient.Conn.Close()
		h.wsClient = nil
	}

	xcontext.Logger(ctx).Infof("User %s unregistered to hub %s (%d)",
		clientID, h.roomID, len(h.clients))

	return nil
}

func (h *hub) ForwardSingleAction(ctx context.Context, action model.GameActionServerRequest) {
	h.pendingRequestMsg <- action
}

func (h *hub) run(ctx context.Context) {
	for {
		isStop := func() bool {
			h.mutex.Lock()
			defer h.mutex.Unlock()

			if !h.isRunning {
				return true
			}

			if h.wsClient != nil {
				return false
			}

			room, err := h.gameRepo.GetRoomByID(ctx, h.roomID)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot get room: %v", err)
				return false
			}

			url := fmt.Sprintf("ws://%s:%s/proxy?room_id=%s",
				room.StartedBy, xcontext.Configs(ctx).GameEngineWSServer.Port, h.roomID)
			conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
			if err != nil {
				xcontext.Logger(ctx).Warnf("Cannot establish a connection with game engine: %v", err)
				return false
			}

			h.wsClient = ws.NewClient(conn)
			go h.runReceive(ctx)
			return false
		}()

		if isStop {
			break
		}

		time.Sleep(5 * time.Second)
	}
}

func (h *hub) runReceive(ctx context.Context) {
	xcontext.Logger(ctx).Infof("Connection to engine of room %s is running", h.roomID)
	for {
		resp, ok := <-h.wsClient.R
		if !ok {
			h.mutex.Lock()
			h.wsClient = nil
			h.mutex.Unlock()
			break
		}

		var actions []model.GameActionServerResponse
		if err := json.Unmarshal(resp, &actions); err != nil {
			xcontext.Logger(ctx).Errorf("Unable to unmarshal: %v", err)
			return
		}

		h.pendingResponseMsg <- actions
	}
	xcontext.Logger(ctx).Infof("Connection to engine of room %s stopped", h.roomID)
}

func (h *hub) broadcast(ctx context.Context) {
	for {
		actions, ok := <-h.pendingResponseMsg
		if !ok {
			break
		}

		for _, action := range actions {
			if err := h.broadcastSingleAction(action); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot broadcast single action to client: %v", err)
				continue
			}
		}
	}
}

func (h *hub) broadcastSingleAction(serverAction model.GameActionServerResponse) error {
	msg, err := json.Marshal(model.ServerActionToClientAction(serverAction))
	if err != nil {
		return err
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()
	if serverAction.To == nil {
		// Broadcast to all clients if the action doesn't specify who will be
		// received the response.
		for _, channel := range h.clients {
			channel <- msg
		}
	} else {
		for _, userID := range serverAction.To {
			channel, ok := h.clients[userID]
			if !ok {
				return err
			}

			channel <- msg
		}
	}

	return nil
}

func (h *hub) runForward(ctx context.Context) {
	ticker := time.NewTicker(xcontext.Configs(ctx).Game.ProxyServerBatchingFrequency)
	defer ticker.Stop()

	batchMsg := []model.GameActionServerRequest{}
	isStop := false
	for !isStop && h.isRunning {
		select {
		case action, ok := <-h.pendingRequestMsg:
			if !ok {
				isStop = true
				break
			}

			batchMsg = append(batchMsg, action)

		case <-ticker.C:
			if len(batchMsg) == 0 {
				continue
			}

			msg, err := json.Marshal(batchMsg)
			batchMsg = batchMsg[:0]
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot marshall batch msg: %v", err)
				continue
			}

			func() {
				h.mutex.Lock()
				defer h.mutex.Unlock()

				if h.wsClient == nil {
					return
				}

				if err := h.wsClient.Write(msg); err != nil {
					xcontext.Logger(ctx).Warnf("Cannot send msg to game engine: %v", err)
					return
				}
			}()
		}
	}
}
