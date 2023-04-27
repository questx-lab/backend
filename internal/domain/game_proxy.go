package domain

import (
	"encoding/json"

	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/internal/domain/gameproxy"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type GameProxyDomain interface {
	ServeGameClient(xcontext.Context, *model.ServeGameClientRequest) error
}

type gameProxyDomain struct {
	gameRepo    repository.GameRepository
	publisher   pubsub.Publisher
	proxyRouter gameproxy.Router
	proxyHubs   *xsync.MapOf[string, gameproxy.Hub]
}

func NewGameProxyDomain(
	gameRepo repository.GameRepository,
	proxyRouter gameproxy.Router,
	publisher pubsub.Publisher,
) GameProxyDomain {
	return &gameProxyDomain{
		gameRepo:    gameRepo,
		publisher:   publisher,
		proxyRouter: proxyRouter,
		proxyHubs:   xsync.NewMapOf[gameproxy.Hub](),
	}
}

func (d *gameProxyDomain) ServeGameClient(ctx xcontext.Context, req *model.ServeGameClientRequest) error {
	userID := xcontext.GetRequestUserID(ctx)
	room, err := d.gameRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return errorx.New(errorx.BadRequest, "Room is not valid")
	}

	hub, _ := d.proxyHubs.LoadOrStore(
		room.ID, gameproxy.NewHub(ctx, ctx.Logger(), d.proxyRouter, d.gameRepo, room.ID))

	// Register client to hub to receive broadcasting messages.
	hubChannel, err := hub.Register(userID)
	if err != nil {
		ctx.Logger().Debugf("Cannot register user to hub: %v", err)
		return errorx.Unknown
	}

	// Join the user in room.
	err = d.publishAction(ctx, room.ID, &gameengine.JoinAction{})
	if err != nil {
		return err
	}

	// Get the initial game state.
	err = d.publishAction(ctx, room.ID, &gameengine.InitAction{})
	if err != nil {
		return err
	}

	defer func() {
		// Remove user from room.
		err = d.publishAction(ctx, room.ID, &gameengine.ExitAction{})
		if err != nil {
			ctx.Logger().Errorf("Cannot create join action: %v", err)
		}

		// Unregister this client from hub.
		err = hub.Unregister(userID)
		if err != nil {
			ctx.Logger().Errorf("Cannot unregister client from hub: %v", err)
		}
	}()

	isStop := false
	for !isStop {
		select {
		case msg, ok := <-ctx.WsClient().R:
			if !ok {
				isStop = true
				break
			}

			clientAction := model.GameActionClientRequest{}
			err := json.Unmarshal(msg, &clientAction)
			if err != nil {
				ctx.Logger().Errorf("Cannot unmarshal client action: %v", err)
				return errorx.Unknown
			}

			serverAction := model.ClientActionToServerAction(clientAction, userID)
			b, err := json.Marshal(serverAction)
			if err != nil {
				ctx.Logger().Errorf("Cannot marshal server action: %v", err)
				return errorx.Unknown
			}

			err = d.publisher.Publish(ctx, model.RequestTopic, &pubsub.Pack{Key: []byte(room.ID), Msg: b})
			if err != nil {
				ctx.Logger().Debugf("Cannot publish action to processor: %v", err)
				return errorx.Unknown
			}

		case msg := <-hubChannel:
			err := ctx.WsClient().Write(msg)
			if err != nil {
				ctx.Logger().Errorf("Cannot write to ws: %v", err)
				return errorx.Unknown
			}
		}
	}

	return nil
}

func (d *gameProxyDomain) publishAction(ctx xcontext.Context, roomID string, action gameengine.Action) error {
	b, err := json.Marshal(model.GameActionServerRequest{
		UserID: xcontext.GetRequestUserID(ctx),
		Type:   action.Type(),
	})
	if err != nil {
		ctx.Logger().Errorf("Cannot marshal action: %v", err)
		return errorx.Unknown
	}

	err = d.publisher.Publish(ctx, model.RequestTopic, &pubsub.Pack{Key: []byte(roomID), Msg: b})
	if err != nil {
		ctx.Logger().Errorf("Cannot publish action: %v", err)
		return errorx.Unknown
	}

	return nil
}
