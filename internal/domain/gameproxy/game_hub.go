package gameproxy

import (
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/log"
	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/domain/gamestate"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type GameHub interface {
	Register(clientID string) (<-chan []byte, error)
	Unregister(clientID string)
	Broadcast(...gamestate.Action)
	Done() <-chan any
}

type gameHub struct {
	done       chan any
	actionChan chan []gamestate.Action

	// clients contains all GameClient registered with this GameHub as keys.
	// If still no action sent to a client, its value is true, otherwise, false.
	clients   *xsync.MapOf[string, chan<- []byte]
	gameState *gamestate.GameState

	newConnectedChan chan string
}

func NewGameHub(ctx xcontext.Context, gameRepo repository.GameRepository, roomID string) (*gameHub, error) {
	gameState, err := gamestate.New(ctx, gameRepo, roomID)
	if err != nil {
		return nil, err
	}

	err = gameState.LoadUser(ctx, gameRepo)
	if err != nil {
		return nil, err
	}

	return &gameHub{
		done:             make(chan any),
		gameState:        gameState,
		actionChan:       make(chan []gamestate.Action),
		clients:          xsync.NewMapOf[chan<- []byte](),
		newConnectedChan: make(chan string),
	}, nil
}

// Broadcast sends a bundle of actions to all clients. This is a non-blocking
// method. After broadcast successfully, an object will be sent to Done channel.
//
// NOTE: Please listen on Done channel to continue broadcasting other actions.
func (hub *gameHub) Broadcast(actionBundle ...gamestate.Action) {
	hub.actionChan <- actionBundle
}

// Register allows a client to subcribe this game hub. All broadcasting actions
// will be sent to this client after this point of time.
func (hub *gameHub) Register(clientID string) (<-chan []byte, error) {
	// To avoid blocking when broadcast to client, we need a bufferred channel
	// here.
	c := make(chan []byte, 1024)

	_, existed := hub.clients.LoadOrStore(clientID, c)
	if existed {
		close(c)
		return nil, errors.New("the game client has already registered")
	}

	hub.newConnectedChan <- clientID
	return c, nil
}

// Unregister removes the game client from this hub.
func (hub *gameHub) Unregister(clientID string) {
	c, existed := hub.clients.LoadAndDelete(clientID)
	if !existed {
		return
	}

	close(c)
}

// Run receives actions from game processor, then broadcasts them to clients.
// This method should be started as a goroutine.
func (hub *gameHub) Run() {
	logger := logger.NewLogger()
	isStop := false

	for !isStop {
		select {
		case actions, ok := <-hub.actionChan:
			if !ok {
				isStop = true
				break
			}

			if len(actions) == 0 {
				hub.done <- nil
				continue
			}

			// Verify actions and update game state. The finalAction will be the
			// concatenation of all serialized actions.
			var finalAction []byte
			for i := range actions {
				actionID, err := hub.gameState.Apply(actions[i])
				if err != nil {
					// Ignore invalid game actions.
					logger.Warnf("Cannot apply game action: %v", err)
					continue
				}

				actionResp, err := gamestate.FormatAction(actionID, actions[i])
				if err != nil {
					logger.Errorf("Cannot format game action: %v", err)
					continue
				}

				b, err := json.Marshal(actionResp)
				if err != nil {
					logger.Warnf("Cannot format action: %v", err)
					continue
				}

				finalAction = append(finalAction, b...)
			}

			if len(finalAction) == 0 {
				hub.done <- nil
				continue
			}

			// Loop over all clients and send finalActions to them.
			hub.clients.Range(func(clientID string, channel chan<- []byte) bool {
				// If the client just connected to the game hub, hub will send it
				// the initial game state.

				channel <- finalAction
				return true
			})

			hub.done <- nil

		case clientID := <-hub.newConnectedChan:
			channel, existed := hub.clients.Load(clientID)
			if !existed {
				logger.Errorf("Not found client id in clients")
				continue
			}

			serializedGameState := hub.gameState.Serialize()
			resp := model.GameActionClientResponse{
				ID:   serializedGameState.ID,
				Type: "init",
				Value: map[string]any{
					"users": serializedGameState.Users,
				},
			}

			b, err := json.Marshal(resp)
			if err != nil {
				log.Error("Cannot marshal initial game state: %v", err)
				continue
			}

			channel <- b
		}
	}

	close(hub.done)
}

func (hub *gameHub) Done() <-chan any {
	return hub.done
}
