package gameengine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/xcontext"
)

const maxPendingActionSize = 1 << 10

type Router interface {
	Register(ctx context.Context, roomID string) (<-chan model.GameActionServerRequest, error)
	Unregister(ctx context.Context, roomID string) error
	Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time)
}

type router struct {
	engineChannels *xsync.MapOf[string, chan<- model.GameActionServerRequest]
}

func NewRouter() Router {
	return &router{
		engineChannels: xsync.NewMapOf[chan<- model.GameActionServerRequest](),
	}
}

func (r *router) Register(ctx context.Context, roomID string) (<-chan model.GameActionServerRequest, error) {
	c := make(chan model.GameActionServerRequest, maxPendingActionSize)
	if _, ok := r.engineChannels.LoadOrStore(roomID, c); ok {
		close(c)
		return nil, errors.New("the room had been registered before")
	}

	return c, nil
}

func (r *router) Unregister(ctx context.Context, roomID string) error {
	roomChannel, ok := r.engineChannels.LoadAndDelete(roomID)
	if !ok {
		return fmt.Errorf("not found room id %s", roomID)
	}

	close(roomChannel)
	return nil
}

func (r *router) Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time) {
	var req model.GameActionServerRequest
	if err := json.Unmarshal(pack.Msg, &req); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to unmarshal: %v", err)
		return
	}

	roomID := string(pack.Key)
	channel, ok := r.engineChannels.Load(roomID)
	if !ok {
		return
	}

	channel <- req
}
