package gameproxy

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/xcontext"
)

const maxMsgSize = 1 << 8

type Hub interface {
	Register(clientID string) (<-chan []byte, error)
	Unregister(clientID string) error
}

type hub struct {
	roomID        string
	router        Router
	isRegistered  bool
	registerMutex sync.Mutex

	logger logger.Logger

	pendingAction <-chan model.GameActionResponse

	// clients contains all GameClient registered with this GameHub as keys.
	// If still no action sent to a client, its value is true, otherwise, false.
	clients *xsync.MapOf[string, chan<- []byte]
}

func NewHub(
	ctx xcontext.Context,
	logger logger.Logger,
	router Router,
	gameRepo repository.GameRepository,
	roomID string,
) *hub {
	hub := &hub{
		isRegistered:  false,
		router:        router,
		registerMutex: sync.Mutex{},
		logger:        logger,
		clients:       xsync.NewMapOf[chan<- []byte](),
		pendingAction: nil,
	}

	go hub.run()

	return hub
}

// Register allows a client to subcribe this game hub. All broadcasting actions
// will be sent to this client after this point of time.
func (h *hub) Register(clientID string) (<-chan []byte, error) {
	h.registerMutex.Lock()
	defer h.registerMutex.Unlock()

	if !h.isRegistered {
		var err error
		h.pendingAction, err = h.router.Register(h.roomID)
		if err != nil {
			return nil, err
		}

		h.isRegistered = true
		go h.run()
	}

	// To avoid blocking when broadcast to client, we need a bufferred channel
	// here.
	c := make(chan []byte, maxMsgSize)

	_, existed := h.clients.LoadOrStore(clientID, c)
	if existed {
		close(c)
		return nil, errors.New("the game client has already registered")
	}

	return c, nil
}

// Unregister removes the game client from this hub.
func (h *hub) Unregister(clientID string) error {
	c, existed := h.clients.LoadAndDelete(clientID)
	if !existed {
		return errors.New("the client has not registered yet")
	}

	close(c)

	h.registerMutex.Lock()
	defer h.registerMutex.Unlock()

	// Temporarily unregister hub from router.
	if h.clients.Size() == 0 {
		if err := h.router.Unregister(h.roomID); err != nil {
			return err
		}
		h.isRegistered = false
	}

	return nil
}

func (h *hub) run() {
	for {
		action, ok := <-h.pendingAction
		if !ok {
			break
		}

		if err := h.broadcast(action); err != nil {
			h.logger.Debugf("Cannot send action bundle to all clients: %v", err)
		}
	}
}

func (h *hub) broadcast(action model.GameActionResponse) error {
	msg, err := json.Marshal(action)
	if err != nil {
		return err
	}

	if action.To == nil {
		// Broadcast to all clients if the action doesn't specify who will be
		// received the response.
		h.clients.Range(func(clientID string, channel chan<- []byte) bool {
			channel <- msg
			return true
		})
	} else {
		for _, userID := range action.To {
			userChannel, ok := h.clients.Load(userID)
			if !ok {
				return err
			}
			userChannel <- msg
		}
	}

	return nil
}
