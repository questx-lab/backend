package gameengine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
)

const maxPendingActionSize = 1 << 10

type Router interface {
	ID() string
	Register(ctx context.Context, roomID string) (<-chan model.GameActionServerRequest, error)
	Unregister(ctx context.Context, roomID string) error
	HandleEvent(ctx context.Context, topic string, pack *pubsub.Pack, t time.Time)
	PingCenter(ctx context.Context)
}

type router struct {
	id            string
	communityRepo repository.CommunityRepository
	gameRepo      repository.GameRepository
	userRepo      repository.UserRepository
	storage       storage.Storage
	publisher     pubsub.Publisher

	engineChannels *xsync.MapOf[string, chan<- model.GameActionServerRequest]
}

func NewRouter(
	communityRepo repository.CommunityRepository,
	gameRepo repository.GameRepository,
	userRepo repository.UserRepository,
	storage storage.Storage,
	publisher pubsub.Publisher,
) Router {
	return &router{
		id:             uuid.NewString(),
		communityRepo:  communityRepo,
		gameRepo:       gameRepo,
		userRepo:       userRepo,
		storage:        storage,
		publisher:      publisher,
		engineChannels: xsync.NewMapOf[chan<- model.GameActionServerRequest](),
	}
}

func (r *router) ID() string {
	return r.id
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

func (r *router) HandleEvent(ctx context.Context, topic string, pack *pubsub.Pack, t time.Time) {
	roomID := string(pack.Key)
	xcontext.Logger(ctx).Infof("topic: %v, key = %v, msg = %v", topic, string(pack.Key), string(pack.Msg))
	switch {
	case len(pack.Msg) > 0:
		var req model.GameActionServerRequest
		if err := json.Unmarshal(pack.Msg, &req); err != nil {
			xcontext.Logger(ctx).Errorf("Unable to unmarshal: %v", err)
			return
		}

		channel, ok := r.engineChannels.Load(roomID)
		if !ok {
			return
		}

		channel <- req

	case len(pack.Msg) == 0:
		_, err := NewEngine(ctx, r, r.publisher, r.gameRepo, r.userRepo, r.storage, roomID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot start game %s: %v", roomID, err)
			return
		}

		xcontext.Logger(ctx).Infof("Start game %s successfully", roomID)
	}
}

func (r *router) PingCenter(ctx context.Context) {
	defer time.AfterFunc(xcontext.Configs(ctx).Game.GameEnginePingFrequency, func() {
		r.PingCenter(ctx)
	})

	err := r.publisher.Publish(ctx, model.GameEnginePingTopic, &pubsub.Pack{Key: []byte(r.id)})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot publish ping topic: %v", err)
		return
	}
}
