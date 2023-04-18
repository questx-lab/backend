package game

import (
	"encoding/json"
	"errors"

	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/domain/gamestate"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type GameHub interface {
	Register(clientID string) (<-chan []byte, error)
	Unregister(clientID string) error
	Broadcast(...gamestate.Action)
	Done() <-chan any
}

type gameClient struct {
	c             chan<- []byte
	justConnected bool
}

type gameHub struct {
	done       chan any
	actionChan chan []gamestate.Action

	// clients contains all GameClient registered with this GameHub as keys.
	// If still no action sent to a client, its value is true, otherwise, false.
	clients   *xsync.MapOf[string, *gameClient]
	gameState *gamestate.GameState
}

func NewGameHub(ctx xcontext.Context, gameRepo repository.GameRepository, roomID string) (*gameHub, error) {
	gameState, err := gamestate.New(ctx, gameRepo, roomID)
	if err != nil {
		return nil, err
	}

	return &gameHub{
		done:       make(chan any),
		gameState:  gameState,
		actionChan: make(chan []gamestate.Action),
		clients:    xsync.NewMapOf[*gameClient](),
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

	_, existed := hub.clients.LoadOrStore(clientID, &gameClient{
		c:             c,
		justConnected: true,
	})

	if existed {
		close(c)
		return nil, errors.New("the game client has already registered")
	}

	return c, nil
}

// Unregister removes the game client from this hub.
func (hub *gameHub) Unregister(clientID string) error {
	client, existed := hub.clients.LoadAndDelete(clientID)
	if !existed {
		return errors.New("the game client has not yet registered to hub")
	}

	close(client.c)
	return nil
}

// Run receives actions from game processor, then broadcasts them to clients.
// This method should be started as a goroutine.
func (hub *gameHub) Run() {
	logger := logger.NewLogger()

	for {
		actions, ok := <-hub.actionChan
		if !ok {
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

		// The initialGameState is only assigned if there is at least one client
		// JUST connected to this hub.
		var initialGameState []byte

		// Loop over all clients and send finalActions to them.
		hub.clients.Range(func(clientID string, client *gameClient) bool {
			// If the client just connected to the game hub, hub will send it
			// the initial game state.
			if client.justConnected {
				if initialGameState == nil {
					var err error
					initialGameState, err = hub.gameState.Serialize()
					if err != nil {
						logger.Errorf("Cannot serialize the game state: %v", err)
						return true
					}
				}

				client.c <- initialGameState
				client.justConnected = false
			}

			client.c <- finalAction
			return true
		})

		hub.done <- nil
	}

	close(hub.done)
}

func (hub *gameHub) Done() <-chan any {
	return hub.done
}
