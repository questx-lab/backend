package gameproxy

import (
	"errors"
	"fmt"

	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/domain/gamestate"
	"github.com/questx-lab/backend/internal/model"
)

type GameRouter interface {
	Register(roomID string) (<-chan gamestate.Action, error)
	Unregister(roomID string) error
	Route(req model.GameActionRouterRequest) error
	Run()
}

type gameRouter struct {
	buffer       chan model.GameActionRouterRequest
	roomChannels *xsync.MapOf[string, chan<- gamestate.Action]
}

func NewGameRouter() *gameRouter {
	return &gameRouter{
		buffer:       make(chan model.GameActionRouterRequest, 1<<16),
		roomChannels: xsync.NewMapOf[chan<- gamestate.Action](),
	}
}

func (router *gameRouter) Register(roomID string) (<-chan gamestate.Action, error) {
	c := make(chan gamestate.Action, 1024)
	if _, existed := router.roomChannels.LoadOrStore(roomID, c); existed {
		close(c)
		return nil, errors.New("the room had been registered before")
	}

	return c, nil
}

func (router *gameRouter) Unregister(roomID string) error {
	roomChannel, existed := router.roomChannels.LoadAndDelete(roomID)
	if !existed {
		return fmt.Errorf("not found room id %s", roomID)
	}

	close(roomChannel)
	return nil
}

func (router *gameRouter) Route(req model.GameActionRouterRequest) error {
	_, existed := router.roomChannels.Load(req.RoomID)
	if !existed {
		return fmt.Errorf("not found room id %s", req.RoomID)
	}

	// request publish
	router.buffer <- req
	return nil
}

// request subscribe
func (router *gameRouter) Run() {
	for {
		actionReq, ok := <-router.buffer
		if !ok {
			break
		}

		roomChannel, existed := router.roomChannels.Load(actionReq.RoomID)
		if !existed {
			continue
		}

		action, err := gamestate.ParseAction(actionReq)
		if err != nil {
			continue
		}

		roomChannel <- action
	}
}
