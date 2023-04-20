package game_v2

import (
	"context"
	"log"
	"time"

	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/puzpuzpuz/xsync"
)

type RoomHub interface {
	Create(ctx xcontext.Context, roomID string) error
	Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time)
}

type roomHub struct {
	roomProcessors *xsync.MapOf[string, RoomProcessor]
	gameRepo       repository.GameRepository
	publisher      pubsub.Publisher
}

func NewRoomHub(
	gameRepo repository.GameRepository,
	publisher pubsub.Publisher,
) RoomHub {
	return &roomHub{
		roomProcessors: xsync.NewMapOf[RoomProcessor](),
		gameRepo:       gameRepo,
	}
}

func (h *roomHub) Create(ctx xcontext.Context, roomID string) error {
	processor, err := NewRoomProcessor(ctx, h.gameRepo, h.publisher, roomID)
	if err != nil {
		return err
	}
	h.roomProcessors.Store(roomID, processor)
	return nil
}

func (h *roomHub) Subscribe(ctx context.Context, pack *pubsub.Pack, t time.Time) {
	roomID := string(pack.Key)
	processor, ok := h.roomProcessors.Load(roomID)
	if !ok {
		log.Printf("Room is not valid: %v", roomID)
		return
	}
	processor.Process(ctx, pack, t)
}
