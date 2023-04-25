package gameproxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/pubsub"
)

const maxPendingActionSize = 1 << 10

type Router interface {
	Register(roomID string) (<-chan model.GameActionResponse, error)
	Unregister(roomID string) error
	Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time)
}

type router struct {
	logger       logger.Logger
	roomChannels *xsync.MapOf[string, chan<- model.GameActionResponse]
}

func NewRouter(logger logger.Logger) *router {
	return &router{
		logger:       logger,
		roomChannels: xsync.NewMapOf[chan<- model.GameActionResponse](),
	}
}

func (r *router) Register(roomID string) (<-chan model.GameActionResponse, error) {
	c := make(chan model.GameActionResponse, maxPendingActionSize)
	if _, ok := r.roomChannels.LoadOrStore(roomID, c); ok {
		close(c)
		return nil, errors.New("the room had been registered before")
	}

	return c, nil
}

func (r *router) Unregister(roomID string) error {
	roomChannel, ok := r.roomChannels.LoadAndDelete(roomID)
	if !ok {
		return fmt.Errorf("not found room id %s", roomID)
	}

	close(roomChannel)
	return nil
}

func (r *router) Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time) {
	var resp model.GameActionResponse
	if err := json.Unmarshal(pack.Msg, &resp); err != nil {
		r.logger.Errorf("Unable to unmarshal: %v", err)
		return
	}

	roomC, ok := r.roomChannels.Load(string(pack.Key))
	if !ok {
		return
	}

	roomC <- resp
}
