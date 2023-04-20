package game_v2

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/questx-lab/backend/internal/domain/gamestate"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync"
)

type RoomProcessor interface {
	Register(clientID string) (<-chan []byte, error)
	Unregister(clientID string)
	Broadcast(...gamestate.Action)
	Process(context.Context, *pubsub.Pack, time.Time)
}

type roomProcessor struct {
	clients   *xsync.MapOf[string, chan<- []byte]
	gameState *gamestate.GameState

	publisher pubsub.Publisher
}

func NewRoomProcessor(
	ctx xcontext.Context,
	gameRepo repository.GameRepository,
	publisher pubsub.Publisher,
	roomID string,
) (*roomProcessor, error) {
	gameState, err := gamestate.New(ctx, gameRepo, roomID)
	if err != nil {
		return nil, err
	}

	err = gameState.LoadUser(ctx, gameRepo)
	if err != nil {
		return nil, err
	}

	return &roomProcessor{
		gameState: gameState,
		clients:   xsync.NewMapOf[chan<- []byte](),
		publisher: publisher,
	}, nil
}

func (h *roomProcessor) Broadcast(actionBundle ...gamestate.Action) {
	if err := h.publisher.Publish(context.Background(), string(model.ResponseTopic), &pubsub.Pack{}); err != nil {
		log.Printf("Unable to publish: %v\n", err)
	}
}

// Register allows a client to subcribe this game hub. All broadcasting actions
// will be sent to this client after this point of time.
func (h *roomProcessor) Register(clientID string) (<-chan []byte, error) {
	c := make(chan []byte, 1024)

	_, existed := h.clients.LoadOrStore(clientID, c)
	if existed {
		close(c)
		return nil, errors.New("the game client has already registered")
	}

	return c, nil
}

// Unregister removes the game client from this hub.
func (hub *roomProcessor) Unregister(clientID string) {
	c, existed := hub.clients.LoadAndDelete(clientID)
	if !existed {
		return
	}

	close(c)
}

func (h *roomProcessor) Process(ctx context.Context, pack *pubsub.Pack, t time.Time) {
	var actionReq model.GameActionRouterRequest
	if err := json.Unmarshal(pack.Msg, &actionReq); err != nil {
		log.Printf("Unable to unmarshal game action router: %v", err)
		return
	}
	action, err := gamestate.ParseAction(actionReq)
	if err != nil {
		log.Printf("Unable to parse action: %v", err)
	}
	actionID, err := h.gameState.Apply(action)
	if err != nil {
		// Ignore invalid game actions.
		log.Printf("Unable to apply game action: %v", err)
		return
	}

	actionResp, err := gamestate.FormatActionV2(actionID, string(pack.Key), action)
	if err != nil {
		log.Printf("Unable to format game action: %v", err)
		return
	}

	b, err := json.Marshal(actionResp)
	if err != nil {
		log.Printf("Unable to format action: %v", err)
		return
	}

	if err := h.publisher.Publish(ctx, string(model.ResponseTopic), &pubsub.Pack{
		Key: []byte(uuid.NewString()),
		Msg: b,
	}); err != nil {
		log.Printf("Unable to publish response: %v", err)
		return
	}
}
