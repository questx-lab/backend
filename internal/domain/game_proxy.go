package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/internal/domain/gameproxy"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type GameProxyDomain interface {
	ServeGameClient(context.Context, *model.ServeGameClientRequest) error
}

type gameProxyDomain struct {
	proxyID           string
	gameRepo          repository.GameRepository
	gameCharacterRepo repository.GameCharacterRepository
	followerRepo      repository.FollowerRepository
	userRepo          repository.UserRepository
	communityRepo     repository.CommunityRepository
	proxyHubs         *xsync.MapOf[string, gameproxy.Hub]
}

func NewGameProxyDomain(
	proxyID string,
	gameRepo repository.GameRepository,
	gameCharacterRepo repository.GameCharacterRepository,
	followerRepo repository.FollowerRepository,
	userRepo repository.UserRepository,
	communityRepo repository.CommunityRepository,
) GameProxyDomain {
	return &gameProxyDomain{
		proxyID:           proxyID,
		gameRepo:          gameRepo,
		gameCharacterRepo: gameCharacterRepo,
		followerRepo:      followerRepo,
		userRepo:          userRepo,
		communityRepo:     communityRepo,
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

	userID := xcontext.RequestUserID(ctx)
	gameUser, err := d.gameRepo.GetUser(ctx, userID, room.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		xcontext.Logger(ctx).Errorf("Cannot get game user: %v", err)
		return errorx.Unknown
	}

	if err == nil && gameUser.ConnectedBy.Valid {
		return errorx.New(errorx.Unavailable, "You're already online or join room to quick, try again after %s",
			xcontext.Configs(ctx).Game.GameSaveFrequency)
	}

	numberUsers, err := d.gameRepo.CountActiveUsersByRoomID(ctx, req.RoomID, userID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot count active users in room: %v", err)
		return errorx.Unknown
	}

	if numberUsers >= int64(xcontext.Configs(ctx).Game.MaxUsers) {
		return errorx.New(errorx.Unavailable, "Room is full")
	}

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
			return err
		}
	}

	hub, ok := d.proxyHubs.LoadOrCompute(room.ID, func() gameproxy.Hub {
		return gameproxy.NewHub(ctx, d.gameRepo, room.ID, d.proxyID)
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
	hub.ForwardSingleAction(ctx, model.GameActionServerRequest{
		UserID: userID,
		Type:   gameengine.JoinAction{}.Type(),
	})

	defer func() {
		// Remove user from room.
		hub.ForwardSingleAction(ctx, model.GameActionServerRequest{
			UserID: userID,
			Type:   gameengine.ExitAction{}.Type(),
		})

		// Unregister this client from hub.
		if err := hub.Unregister(ctx, userID); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot unregister client from hub: %v", err)
		}

		err := d.gameRepo.UpsertGameUserWithProxy(ctx, &entity.GameUser{
			UserID:      userID,
			RoomID:      req.RoomID,
			ConnectedBy: sql.NullString{},
		})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot update proxy of user: %v", err)
		}
	}()

	wsClient := xcontext.WSClient(ctx)
	isStop := false
	lastMoveAction := time.Now()
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

			if clientAction.Type == (gameengine.MoveAction{}).Type() {
				if time.Since(lastMoveAction) < 10*time.Millisecond {
					continue
				}

				lastMoveAction = time.Now()
			}

			go hub.ForwardSingleAction(ctx, model.ClientActionToServerAction(clientAction, userID))

		case msg := <-hubChannel:
			if size := len(hubChannel); size > 50 {
				xcontext.Logger(ctx).Errorf("Bottleneck detected when sending msg to client, ratio=%d",
					size)
			}

			if err := wsClient.Write(msg); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot write to ws: %v", err)
				return errorx.Unknown
			}
		}
	}

	return nil
}
