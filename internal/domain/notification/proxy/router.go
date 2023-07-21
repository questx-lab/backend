package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/questx-lab/backend/internal/domain/notification/directive"
	"github.com/questx-lab/backend/internal/domain/notification/event"
	"github.com/questx-lab/backend/pkg/ws"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type Router struct {
	engineClient  *ws.Client
	communityHubs map[string]*CommunityHub
	userHubs      map[string]*UserHub

	mutex sync.RWMutex
}

func NewRouter(ctx context.Context) *Router {
	router := &Router{
		engineClient:  nil,
		communityHubs: make(map[string]*CommunityHub),
		userHubs:      make(map[string]*UserHub),
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
		r.checkConnection(ctx)
		if err := r.cleanup(ctx); err != nil {
			xcontext.Logger(ctx).Warnf("An error occurred when clean up router: %v", err)
		}

		time.Sleep(5000 * time.Millisecond)
	}
}

func (r *Router) cleanup(ctx context.Context) error {
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
		return nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.engineClient == nil {
		return nil
	}

	for _, communityID := range emptyCommunityHubs {
		if _, ok := r.communityHubs[communityID]; ok {
			b, err := json.Marshal(directive.NewUnregisterCommunityDirective(communityID))
			if err != nil {
				return err
			}

			if err := r.engineClient.Write(b, true); err != nil {
				return err
			}

			close(r.communityHubs[communityID].c)
			delete(r.communityHubs, communityID)
		}
	}

	for _, userID := range emptyUserHubs {
		if _, ok := r.userHubs[userID]; ok {
			b, err := json.Marshal(directive.NewUnregisterUserDirective(userID))
			if err != nil {
				return err
			}

			if err := r.engineClient.Write(b, true); err != nil {
				return err
			}

			delete(r.userHubs, userID)
		}
	}

	return nil
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
		if event.Metadata.ToCommunity != "" {
			hub, ok := r.communityHubs[event.Metadata.ToCommunity]
			if ok {
				hub.c <- &event
			}
		} else {
			hub, ok := r.userHubs[event.Metadata.ToUser]
			if ok {
				go hub.Send(&event)
			}
		}
		r.mutex.RUnlock()
	}
}
