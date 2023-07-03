package domain

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/internal/domain/gameproxy"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/buffer"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type GameProxyDomain interface {
	ServeGameClient(context.Context, *model.ServeGameClientRequest) error
}

type gameProxyDomain struct {
	gameRepo          repository.GameRepository
	gameCharacterRepo repository.GameCharacterRepository
	followerRepo      repository.FollowerRepository
	userRepo          repository.UserRepository
	communityRepo     repository.CommunityRepository
	publisher         pubsub.Publisher
	proxyRouter       gameproxy.Router
	proxyHubs         *xsync.MapOf[string, gameproxy.Hub]
}

func NewGameProxyDomain(
	gameRepo repository.GameRepository,
	gameCharacterRepo repository.GameCharacterRepository,
	followerRepo repository.FollowerRepository,
	userRepo repository.UserRepository,
	communityRepo repository.CommunityRepository,
	proxyRouter gameproxy.Router,
	publisher pubsub.Publisher,
) GameProxyDomain {
	return &gameProxyDomain{
		gameRepo:          gameRepo,
		gameCharacterRepo: gameCharacterRepo,
		followerRepo:      followerRepo,
		userRepo:          userRepo,
		communityRepo:     communityRepo,
		publisher:         publisher,
		proxyRouter:       proxyRouter,
		proxyHubs:         xsync.NewMapOf[gameproxy.Hub](),
	}
}

func (d *gameProxyDomain) ServeGameClient(ctx context.Context, req *model.ServeGameClientRequest) error {
	room, err := d.gameRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get room: %v", err)
		return errorx.New(errorx.BadRequest, "Room is not valid")
	}

	if room.StartedBy == "" {
		return errorx.New(errorx.Unavailable, "Room has not started yet")
	}

	numberUsers, err := d.gameRepo.CountActiveUsersByRoomID(ctx, req.RoomID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot count active users in room: %v", err)
		return errorx.Unknown
	}

	if numberUsers >= int64(xcontext.Configs(ctx).Game.MaxUsers) {
		return errorx.New(errorx.Unavailable, "Room is full")
	}

	userID := xcontext.RequestUserID(ctx)
	userCharacter, err := d.gameCharacterRepo.GetAllUserCharacters(ctx, userID, room.CommunityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get user characters: %v", err)
		return errorx.Unknown
	}

	if len(userCharacter) == 0 {
		return errorx.New(errorx.Unavailable, "User must buy a character before")
	}

	// Check if user follows the community.
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

	hub, ok := d.proxyHubs.LoadOrCompute(room.ID, func() gameproxy.Hub {
		return gameproxy.NewHub(ctx, d.proxyRouter, d.gameRepo, room.ID)
	})

	// When use LoadOrCompute, the returned object and stored object in the
	// first time are difference. So when the first user joins in the room,
	// others cannot join because the hub registered at the first time and the
	// hub which other users join later are not the same. Until the first user
	// leaves the room, the room will return to the normal operation.
	// So at the first time, we need to "re"-load the hub again to make sure the
	// returned hub is the stored one in the map.
	if !ok {
		hub, _ = d.proxyHubs.Load(room.ID)
	}

	// Register client to hub to receive broadcasting messages.
	hubChannel, err := hub.Register(ctx, userID)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Cannot register user to hub: %v", err)
		return errorx.New(errorx.Unavailable, "You have already joined in room")
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
		err = hub.Unregister(ctx, userID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot unregister client from hub: %v", err)
		}
	}()

	wsClient := xcontext.WSClient(ctx)

	var pendingMsg [][]byte
	var ticker = time.NewTicker(xcontext.Configs(ctx).Game.ProxyBatchingFrequency)

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
			pendingMsg = append(pendingMsg, msg)

		case <-ticker.C:
			if len(pendingMsg) == 0 {
				continue
			}

			buf := buffer.New()
			buf.AppendByte('[')

			for i, msg := range pendingMsg {
				buf.AppendBytes(msg)
				if i < len(pendingMsg)-1 {
					buf.AppendByte(',')
				}
			}

			buf.AppendByte(']')
			pendingMsg = pendingMsg[:0]

			err := wsClient.Write(buf.Bytes())
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
