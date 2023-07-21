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
	engineClient *ws.Client
	hubs         map[string]*Hub

	mutex sync.RWMutex
}

func NewRouter(ctx context.Context) *Router {
	router := &Router{
		engineClient: nil,
		hubs:         make(map[string]*Hub),
		mutex:        sync.RWMutex{},
	}

	go router.run(ctx)
	return router
}

func (r *Router) GetHub(ctx context.Context, communityID string) (*Hub, error) {
	r.mutex.RLock()
	hub, ok := r.hubs[communityID]
	r.mutex.RUnlock()
	if ok {
		return hub, nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.hubs[communityID]; !ok {
		b, err := json.Marshal(directive.NewRegisterCommunityDirective(communityID))
		if err != nil {
			return nil, err
		}

		if r.engineClient == nil {
			return nil, errors.New("not found hub")
		}

		if err := r.engineClient.Write(b, true); err != nil {
			return nil, err
		}

		r.hubs[communityID] = NewHub(communityID)
		xcontext.Logger(ctx).Infof("Registered to community %s successfully", communityID)
	}

	return r.hubs[communityID], nil
}

func (r *Router) run(ctx context.Context) {
	for {
		r.checkConnection(ctx)
		r.cleanup(ctx)
		time.Sleep(5000 * time.Millisecond)
	}
}

func (r *Router) cleanup(ctx context.Context) error {
	emptyHubs := []string{}

	r.mutex.RLock()
	for _, h := range r.hubs {
		if h.IsEmpty() {
			emptyHubs = append(emptyHubs, h.communityID)
		}
	}
	r.mutex.RUnlock()

	if len(emptyHubs) == 0 {
		return nil
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, communityID := range emptyHubs {
		if _, ok := r.hubs[communityID]; ok {
			b, err := json.Marshal(directive.NewUnregisterCommunityDirective(communityID))
			if err != nil {
				return err
			}

			if err := r.engineClient.Write(b, true); err != nil {
				return err
			}

			close(r.hubs[communityID].c)
			delete(r.hubs, communityID)
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
	r.mutex.Lock()
	for _, h := range r.hubs {
		b, err := json.Marshal(directive.NewRegisterCommunityDirective(h.communityID))
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot marshal directive: %v", err)
			continue
		}

		if err := r.engineClient.Write(b, true); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot register hub %s to engine: %v", h.communityID, err)
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
		hub, ok := r.hubs[event.Metadata.To]
		if ok {
			hub.c <- &event
		}
		r.mutex.RUnlock()
	}
}
