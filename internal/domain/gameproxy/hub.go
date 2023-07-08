package gameproxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/ws"
	"github.com/questx-lab/backend/pkg/xcontext"
)

const maxMsgSize = 1 << 13

type Hub interface {
	Register(ctx context.Context, clientID string) (<-chan []byte, error)
	Unregister(ctx context.Context, clientID string) error
	ForwardSingleAction(ctx context.Context, action model.GameActionServerRequest)
}

type hub struct {
	roomID  string
	proxyID string

	pendingResponseActions chan model.GameActionServerResponse
	pendingRequestActions  chan model.GameActionServerRequest

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
	proxyID string,
) *hub {
	hub := &hub{
		roomID:  roomID,
		proxyID: proxyID,

		gameRepo:               gameRepo,
		pendingResponseActions: make(chan model.GameActionServerResponse, maxMsgSize),
		pendingRequestActions:  make(chan model.GameActionServerRequest, maxMsgSize),

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
		xcontext.Logger(ctx).Infof("Prepare to close connection to engine %s", h.roomID)
	}

	xcontext.Logger(ctx).Infof("User %s unregistered to hub %s (%d)",
		clientID, h.roomID, len(h.clients))

	return nil
}

func (h *hub) ForwardSingleAction(ctx context.Context, action model.GameActionServerRequest) {
	h.pendingRequestActions <- action
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

			url := fmt.Sprintf("ws://%s:%s/proxy?room_id=%s&proxy_id=%s",
				room.StartedBy, xcontext.Configs(ctx).GameEngineWSServer.Port, h.roomID, h.proxyID)
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

		go func() {
			for _, action := range actions {
				h.pendingResponseActions <- action
			}
		}()
	}
	xcontext.Logger(ctx).Infof("Connection to engine of room %s stopped", h.roomID)
}

func (h *hub) broadcast(ctx context.Context) {
	broadcastActions := []model.GameActionClientResponse{}
	ticker := time.NewTicker(xcontext.Configs(ctx).Game.ProxyClientBatchingFrequency)
	defer ticker.Stop()

	for {
		select {
		case action, ok := <-h.pendingResponseActions:
			if !ok {
				break
			}

			if action.To == nil {
				broadcastActions = append(broadcastActions, model.ServerActionToClientAction(action))
			} else {
				go func() {
					if err := h.sendSingleAction(action); err != nil {
						xcontext.Logger(ctx).Errorf("Cannot send single action to client: %v", err)
					}
				}()
			}

		case <-ticker.C:
			if len(broadcastActions) == 0 {
				continue
			}

			if err := h.broadcastActions(broadcastActions); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot broadcast batch of actions to client: %v", err)
			}

			broadcastActions = broadcastActions[:0]
		}
	}
}

func (h *hub) broadcastActions(clientActions []model.GameActionClientResponse) error {
	msg, err := json.Marshal(clientActions)
	if err != nil {
		return err
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()
	for _, channel := range h.clients {
		channel <- msg
	}

	return nil
}

func (h *hub) sendSingleAction(serverAction model.GameActionServerResponse) error {
	msg, err := json.Marshal(
		[]model.GameActionClientResponse{model.ServerActionToClientAction(serverAction)})
	if err != nil {
		return err
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()
	for _, userID := range serverAction.To {
		channel, ok := h.clients[userID]
		if !ok {
			return errors.New("not found user connection in proxy")
		}
		channel <- msg
	}

	return nil
}

func (h *hub) runForward(ctx context.Context) {
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	pendingActions := []model.GameActionServerRequest{}
	isStop := false
	for !isStop && h.isRunning {
		select {
		case action, ok := <-h.pendingRequestActions:
			if !ok {
				isStop = true
				break
			}

			pendingActions = append(pendingActions, action)

		case <-ticker.C:
			if len(pendingActions) == 0 {
				continue
			}

			batch := common.Batch(&pendingActions, 1024)
			common.DetectBottleneck(ctx, batch, pendingActions, "sending requests to engine")

			msg, err := json.Marshal(batch)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot marshall batch msg: %v", err)
				continue
			}

			func() {
				h.mutex.Lock()
				defer h.mutex.Unlock()

				if h.wsClient != nil {
					err := h.wsClient.Write(msg)
					if err == nil {
						return
					}
					xcontext.Logger(ctx).Warnf("Cannot send msg to game engine: %v", err)
				}

				// When we can't connect to engine. We only keep non-move
				// message.
				pendingActions = append(batch, pendingActions...)
				newBatch := []model.GameActionServerRequest{}
				for _, a := range pendingActions {
					if a.Type != (gameengine.MoveAction{}).Type() {
						newBatch = append(newBatch, a)
					}
				}

				pendingActions = newBatch
				xcontext.Logger(ctx).Warnf("Ignore all move actions, pending requests=%d", len(pendingActions))
			}()
		}
	}
}
