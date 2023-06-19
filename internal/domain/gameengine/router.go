package gameengine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
)

const maxPendingActionSize = 1 << 10

type Router interface {
	Register(ctx context.Context, roomID string) (<-chan model.GameActionServerRequest, error)
	Unregister(ctx context.Context, roomID string) error
	Subscribe(ctx context.Context, topic string, pack *pubsub.Pack, t time.Time)
}

type router struct {
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
		communityRepo:  communityRepo,
		gameRepo:       gameRepo,
		userRepo:       userRepo,
		storage:        storage,
		publisher:      publisher,
		engineChannels: xsync.NewMapOf[chan<- model.GameActionServerRequest](),
	}
}

func (r *router) StartCurrentGames(ctx context.Context) error {
	rooms, err := r.gameRepo.GetRoomsByCommunityID(ctx, "")
	if err != nil {
		return err
	}

	for _, room := range rooms {
		_, err := NewEngine(ctx, r, r.publisher, r.gameRepo, r.userRepo, r.storage, room.ID)
		if err != nil {
			return err
		}
	}

	return nil
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

func (r *router) Subscribe(ctx context.Context, topic string, pack *pubsub.Pack, t time.Time) {
	switch topic {
	case model.GameActionRequestTopic:
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
	case model.CreateCommunityTopic:
		communityID := string(pack.Key)
		community, err := r.communityRepo.GetByID(ctx, communityID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Not found community id %s: %v", communityID, err)
			return
		}

		firstMap, err := r.gameRepo.GetFirstMap(ctx)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Not found the first map in db: %v", err)
			return
		}

		room := entity.GameRoom{
			Base:        entity.Base{ID: uuid.NewString()},
			CommunityID: communityID,
			MapID:       firstMap.ID,
			Name:        fmt.Sprintf("%s-%d", community.Handle, crypto.RandRange(100, 999)),
		}
		if err := r.gameRepo.CreateRoom(ctx, &room); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot create room for %s: %v", community.Handle, err)
			return
		}

		_, err = NewEngine(ctx, r, r.publisher, r.gameRepo, r.userRepo, r.storage, room.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot start game for %s: %v", community.Handle, err)
			return
		}

		xcontext.Logger(ctx).Infof("Start game for %s successfully", community.Handle)
	}
}
