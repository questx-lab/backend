package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/notification/directive"
	"github.com/questx-lab/backend/internal/domain/notification/event"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/ws"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
)

type Router struct {
	followerRepo repository.FollowerRepository

	engineClient  *ws.Client
	communityHubs map[string]*CommunityHub
	userHubs      map[string]*UserHub

	engineCaller client.NotificationEngineCaller
	redisClient  xredis.Client
	mutex        sync.RWMutex
}

func NewRouter(
	ctx context.Context,
	followerRepo repository.FollowerRepository,
	engineCaller client.NotificationEngineCaller,
	redisClient xredis.Client,
) *Router {
	router := &Router{
		followerRepo:  followerRepo,
		engineClient:  nil,
		communityHubs: make(map[string]*CommunityHub),
		userHubs:      make(map[string]*UserHub),
		engineCaller:  engineCaller,
		redisClient:   redisClient,
		mutex:         sync.RWMutex{},
	}

	go router.run(ctx)
	return router
}

func (r *Router) GetCommunityHub(ctx context.Context, communityID string) (*CommunityHub, error) {
	r.mutex.RLock()
	hub, ok := r.communityHubs[communityID]
	r.mutex.RUnlock()
	if ok {
		return hub, nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.communityHubs[communityID]; !ok {
		if r.engineClient == nil {
			return nil, errors.New("engine has not started")
		}

		b, err := json.Marshal(directive.NewRegisterCommunityDirective(communityID))
		if err != nil {
			return nil, err
		}
		if err := r.engineClient.Write(b, true); err != nil {
			return nil, err
		}

		r.communityHubs[communityID] = NewCommunityHub(communityID)
		xcontext.Logger(ctx).Infof("Registered to community %s successfully", communityID)
	}

	return r.communityHubs[communityID], nil
}

func (r *Router) GetUserHub(ctx context.Context, userID string) (*UserHub, error) {
	r.mutex.RLock()
	hub, ok := r.userHubs[userID]
	r.mutex.RUnlock()
	if ok {
		return hub, nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.userHubs[userID]; !ok {
		if r.engineClient == nil {
			return nil, errors.New("engine has not started")
		}

		b, err := json.Marshal(directive.NewRegisterUserDirective(userID))
		if err != nil {
			return nil, err
		}
		if err := r.engineClient.Write(b, true); err != nil {
			return nil, err
		}

		r.userHubs[userID] = NewUserHub(userID)
		xcontext.Logger(ctx).Infof("Registered to user %s successfully", userID)
	}

	return r.userHubs[userID], nil
}

func (r *Router) run(ctx context.Context) {
	for {
		log.Println("check connection")
		r.checkConnection(ctx)
		log.Println("cleanup")
		r.cleanup(ctx)
		log.Println("ping status")
		r.pingUserStatus(ctx)

		log.Println("RUN CYCLE")
		time.Sleep(15 * time.Second)
		log.Println("DONE WAIT CYCLE")
	}
}

func (r *Router) cleanup(ctx context.Context) {
	emptyCommunityHubs := []string{}
	emptyUserHubs := []string{}

	r.mutex.RLock()
	for _, h := range r.communityHubs {
		if h.IsEmpty() {
			emptyCommunityHubs = append(emptyCommunityHubs, h.communityID)
		}
	}
	for _, h := range r.userHubs {
		if h.IsEmpty() {
			emptyUserHubs = append(emptyUserHubs, h.userID)
		}
	}
	r.mutex.RUnlock()

	if len(emptyCommunityHubs) == 0 && len(emptyUserHubs) == 0 {
		return
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.engineClient == nil {
		return
	}

	for _, communityID := range emptyCommunityHubs {
		if _, ok := r.communityHubs[communityID]; ok {
			b, err := json.Marshal(directive.NewUnregisterCommunityDirective(communityID))
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot marshal register community directive: %v", err)
				return
			}

			if err := r.engineClient.Write(b, true); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot send register community directive: %v", err)
				return
			}

			delete(r.communityHubs, communityID)
		}
	}

	for _, userID := range emptyUserHubs {
		if _, ok := r.userHubs[userID]; ok {
			b, err := json.Marshal(directive.NewUnregisterUserDirective(userID))
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot marshal register user directive: %v", err)
				return
			}

			if err := r.engineClient.Write(b, true); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot send register user directive: %v", err)
				return
			}

			delete(r.userHubs, userID)
		}
	}
}

func (r *Router) checkConnection(ctx context.Context) {
	r.mutex.RLock()
	engineClient := r.engineClient
	r.mutex.RUnlock()

	if engineClient != nil {
		return
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Double check.
	if r.engineClient != nil {
		return
	}

	url := xcontext.Configs(ctx).Notification.EngineWSServer.Endpoint
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot establish connection with chat engine: %v", err)
		return
	}

	xcontext.Logger(ctx).Infof("Reconnect to notification engine succesfully")

	r.engineClient = ws.NewClient(conn)
	go r.runReceive(ctx)
}

func (r *Router) runReceive(ctx context.Context) {
	// Register all communities and users to engine.
	r.mutex.Lock()
	for _, c := range r.communityHubs {
		b, err := json.Marshal(directive.NewRegisterCommunityDirective(c.communityID))
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot marshal directive: %v", err)
			continue
		}

		if err := r.engineClient.Write(b, true); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot register community %s to engine: %v", c.communityID, err)
			continue
		}
	}

	for _, u := range r.userHubs {
		b, err := json.Marshal(directive.NewRegisterUserDirective(u.userID))
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot marshal directive: %v", err)
			continue
		}

		if err := r.engineClient.Write(b, true); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot register user %s to engine: %v", u.userID, err)
			continue
		}
	}
	r.mutex.Unlock()

	for {
		eventResp, ok := <-r.engineClient.R
		if !ok {
			r.mutex.Lock()
			r.engineClient = nil
			r.mutex.Unlock()
			break
		}

		var event event.EventRequest
		if err := json.Unmarshal(eventResp, &event); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot unmarshal event: %v", err)
			continue
		}

		r.mutex.RLock()
		for _, communityID := range event.Metadata.ToCommunities {
			hub, ok := r.communityHubs[communityID]
			if ok {
				// We send the event to a long-live goroutine to broadcast to
				// users. Because the number of communities is small, and they
				// are always active, so one long-live goroutine is faster than
				// we start a ephemeral goroutine every an event is emited.
				hub.c <- &event
			}
		}

		for _, userID := range event.Metadata.ToUsers {
			hub, ok := r.userHubs[userID]
			if ok {
				// Because the number of users is too large, and no much event
				// send directly to a user, so each long-live goroutine for one
				// user is expensive.
				go hub.Send(&event)
			}
		}
		r.mutex.RUnlock()
	}
}

func (r *Router) pingUserStatus(ctx context.Context) {
	r.mutex.RLock()
	userIDs := common.MapKeys(r.userHubs)
	r.mutex.RUnlock()

	if len(userIDs) == 0 {
		return
	}

	now := time.Now().Unix()
	pingMap := map[string]any{}
	for _, userID := range userIDs {
		pingMap[common.RedisKeyUserStatus(userID)] = now
	}

	keys := common.MapKeys(pingMap)
	values, err := r.redisClient.MGet(ctx, keys...)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get current user status from redis: %v", err)
		return
	}

	if err := r.redisClient.MSet(ctx, pingMap); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot ping user status to redis: %v", err)
	}

	for i := range keys {
		if values[i] == nil { // This key is never set or cleaned up before.
			userID := common.FromRedisKeyUserStatus(keys[i])
			followers, err := r.followerRepo.GetListByUserID(ctx, userID)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot get followers: %v", err)
				continue
			}

			for _, f := range followers {
				err := r.redisClient.SAdd(ctx, common.RedisKeyCommunityOnline(f.CommunityID), userID)
				if err != nil {
					xcontext.Logger(ctx).Errorf("Cannot add member to community online: %v", err)
					continue
				}
			}

			communityIDs := []string{}
			for _, f := range followers {
				communityIDs = append(communityIDs, f.CommunityID)
			}

			ev := event.New(
				event.ChangeUserStatusEvent{UserID: userID, Status: event.Online},
				&event.Metadata{ToCommunities: communityIDs},
			)
			if err := r.engineCaller.Emit(ctx, ev); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot emit online event: %v", err)
				continue
			}
		}
	}
}
