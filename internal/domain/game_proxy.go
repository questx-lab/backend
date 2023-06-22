package domain

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/internal/domain/gameproxy"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type GameProxyDomain interface {
	ServeGameClient(context.Context, *model.ServeGameClientRequest) error
}

type gameProxyDomain struct {
	gameRepo      repository.GameRepository
	followerRepo  repository.FollowerRepository
	userRepo      repository.UserRepository
	communityRepo repository.CommunityRepository
	publisher     pubsub.Publisher
	proxyRouter   gameproxy.Router
	proxyHubs     *xsync.MapOf[string, gameproxy.Hub]
}

func NewGameProxyDomain(
	gameRepo repository.GameRepository,
	followerRepo repository.FollowerRepository,
	userRepo repository.UserRepository,
	communityRepo repository.CommunityRepository,
	proxyRouter gameproxy.Router,
	publisher pubsub.Publisher,
) GameProxyDomain {
	return &gameProxyDomain{
		gameRepo:      gameRepo,
		followerRepo:  followerRepo,
		userRepo:      userRepo,
		communityRepo: communityRepo,
		publisher:     publisher,
		proxyRouter:   proxyRouter,
		proxyHubs:     xsync.NewMapOf[gameproxy.Hub](),
	}
}

func (d *gameProxyDomain) ServeGameClient(ctx context.Context, req *model.ServeGameClientRequest) error {
	room, err := d.gameRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get room: %v", err)
		return errorx.New(errorx.BadRequest, "Room is not valid")
	}

	numberUsers, err := d.gameRepo.CountActiveUsersByRoomID(ctx, req.RoomID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot count active users in room: %v", err)
		return errorx.Unknown
	}

	if numberUsers >= int64(xcontext.Configs(ctx).Game.MaxUsers) {
		return errorx.New(errorx.Unavailable, "Room is full")
	}

	// Check if user follows the community.
	userID := xcontext.RequestUserID(ctx)
	_, err = d.followerRepo.Get(ctx, userID, room.CommunityID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			xcontext.Logger(ctx).Errorf("Cannot get the follower: %v", err)
			return errorx.Unknown
		}

		err := followCommunity(
			ctx,
			d.userRepo,
			d.communityRepo,
			d.followerRepo,
			nil,
			userID, room.CommunityID, "",
		)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot auto follow community: %v", err)
			return err
		}
	}

	hub, _ := d.proxyHubs.LoadOrStore(
		room.ID, gameproxy.NewHub(ctx, xcontext.Logger(ctx), d.proxyRouter, d.gameRepo, room.ID))

	// Register client to hub to receive broadcasting messages.
	hubChannel, err := hub.Register(userID)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot register user to hub: %v", err)
		return errorx.Unknown
	}

	// Join the user in room.
	err = d.publishAction(ctx, room.ID, room.StartedBy, &gameengine.JoinAction{})
	if err != nil {
		return err
	}

	// Get the initial game state.
	err = d.publishAction(ctx, room.ID, room.StartedBy, &gameengine.InitAction{})
	if err != nil {
		return err
	}

	defer func() {
		// Remove user from room.
		err = d.publishAction(ctx, room.ID, room.StartedBy, &gameengine.ExitAction{})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot create join action: %v", err)
		}

		// Unregister this client from hub.
		err = hub.Unregister(userID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot unregister client from hub: %v", err)
		}
	}()

	wsClient := xcontext.WSClient(ctx)
	isStop := false
	for !isStop {
		select {
		case msg, ok := <-wsClient.R:
			if !ok {
				isStop = true
				break
			}

			clientAction := model.GameActionClientRequest{}
			err := json.Unmarshal(msg, &clientAction)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot unmarshal client action: %v", err)
				return errorx.Unknown
			}

			serverAction := model.ClientActionToServerAction(clientAction, userID)
			b, err := json.Marshal(serverAction)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot marshal server action: %v", err)
				return errorx.Unknown
			}

			err = d.publisher.Publish(ctx, room.StartedBy, &pubsub.Pack{Key: []byte(room.ID), Msg: b})
			if err != nil {
				xcontext.Logger(ctx).Debugf("Cannot publish action to processor: %v", err)
				return errorx.Unknown
			}

		case msg := <-hubChannel:
			err := wsClient.Write(msg)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot write to ws: %v", err)
				return errorx.Unknown
			}
		}
	}

	return nil
}

func (d *gameProxyDomain) publishAction(ctx context.Context, roomID, engineID string, action gameengine.Action) error {
	b, err := json.Marshal(model.GameActionServerRequest{
		UserID: xcontext.RequestUserID(ctx),
		Type:   action.Type(),
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot marshal action: %v", err)
		return errorx.Unknown
	}
	err = d.publisher.Publish(ctx, engineID, &pubsub.Pack{Key: []byte(roomID), Msg: b})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot publish action: %v", err)
		return errorx.Unknown
	}

	return nil
}
