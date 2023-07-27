package engine

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/questx-lab/backend/internal/domain/notification/directive"
	"github.com/questx-lab/backend/internal/domain/notification/event"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type EngineServer struct {
	communityProcessor map[string]*CommunityProcessor
	userProcessors     map[string]*UserProcessor
	mutex              sync.RWMutex
}

func NewEngineServer() *EngineServer {
	return &EngineServer{
		communityProcessor: make(map[string]*CommunityProcessor),
		userProcessors:     make(map[string]*UserProcessor),
		mutex:              sync.RWMutex{},
	}
}

func (s *EngineServer) GetCommunityProcessor(communityID string, createIfNotExist bool) *CommunityProcessor {
	s.mutex.RLock()
	community, ok := s.communityProcessor[communityID]
	s.mutex.RUnlock()

	if ok {
		return community
	}

	if !createIfNotExist {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Double check.
	if community, ok := s.communityProcessor[communityID]; ok {
		return community
	}

	s.communityProcessor[communityID] = NewCommunityProcessor(communityID)
	return s.communityProcessor[communityID]
}

func (s *EngineServer) GetUserProcessor(userID string, createIfNotExist bool) *UserProcessor {
	s.mutex.RLock()
	userSession, ok := s.userProcessors[userID]
	s.mutex.RUnlock()

	if ok {
		return userSession
	}

	if !createIfNotExist {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Double check.
	if userProcessor, ok := s.userProcessors[userID]; ok {
		return userProcessor
	}

	s.userProcessors[userID] = NewUserProcessor(userID)
	return s.userProcessors[userID]
}

// Emit handles a emit call from client. It broadcasts the event to every proxy
// registered to the community.
func (s *EngineServer) Emit(_ context.Context, event *event.EventRequest) error {
	start := time.Now()
	if event.Metadata.ToCommunity != "" {
		processor := s.GetCommunityProcessor(event.Metadata.ToCommunity, false)
		if processor != nil {
			processor.Broadcast(event)
		}
	} else {
		processor := s.GetUserProcessor(event.Metadata.ToUser, false)
		if processor != nil {
			processor.Send(event)
		}
	}
	log.Println("BROADCAST", time.Since(start))
	return nil
}

func (s *EngineServer) ServeProxy(ctx context.Context, _ *model.ServeNotificationEngineRequest) error {
	proxySession := NewProxySession()
	defer proxySession.Leave()

	wsClient := xcontext.WSClient(ctx)
	for {
		select {
		case req, ok := <-wsClient.R:
			if !ok {
				return errorx.Unknown
			}

			var d directive.ServerDirective
			if err := json.Unmarshal(req, &d); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot unmarshal directive: %v", err)
				return errorx.Unknown
			}

			switch d.Op {
			case directive.EngineRegisterCommunityDirectiveOp:
				var registerDirective directive.EngineRegisterCommunityDirective
				if err := json.Unmarshal(d.Data, &registerDirective); err != nil {
					xcontext.Logger(ctx).Errorf("Cannot unmarshal register community data: %v", err)
					return errorx.Unknown
				}

				communityProcessor := s.GetCommunityProcessor(registerDirective.CommunityID, true)
				proxySession.RegisterCommunity(communityProcessor)

				xcontext.Logger(ctx).Infof("Proxy %s register to community %s",
					proxySession.id, registerDirective.CommunityID)

			case directive.EngineUnregisterCommunityDirectiveOp:
				var unregisterDirective directive.EngineUnregisterCommunityDirective
				if err := json.Unmarshal(d.Data, &unregisterDirective); err != nil {
					xcontext.Logger(ctx).Errorf("Cannot unmarshal unregister community data: %v", err)
					return errorx.Unknown
				}

				communityProcessor := s.GetCommunityProcessor(unregisterDirective.CommunityID, false)
				proxySession.UnregisterCommunity(communityProcessor)
				xcontext.Logger(ctx).Infof("Proxy %s unregister to community %s",
					proxySession.id, unregisterDirective.CommunityID)

			case directive.EngineRegisterUserDirectiveOp:
				var registerUserDirective directive.EngineRegisterUserDirective
				if err := json.Unmarshal(d.Data, &registerUserDirective); err != nil {
					xcontext.Logger(ctx).Errorf("Cannot unmarshal unregister user data: %v", err)
					return errorx.Unknown
				}

				userSession := s.GetUserProcessor(registerUserDirective.UserID, true)
				if userSession != nil {
					proxySession.RegisterUser(userSession)
				}

			case directive.EngineUnregisterUserDirectiveOp:
				var unregisterUserDirective directive.EngineUnregisterUserDirective
				if err := json.Unmarshal(d.Data, &unregisterUserDirective); err != nil {
					xcontext.Logger(ctx).Errorf("Cannot unmarshal unregister user data: %v", err)
					return errorx.Unknown
				}

				userSession := s.GetUserProcessor(unregisterUserDirective.UserID, false)
				if userSession != nil {
					proxySession.UnregisterUser(userSession)
				}
			}

		case ev, ok := <-proxySession.C:
			if !ok {
				return errorx.Unknown
			}

			b, err := json.Marshal(ev)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot marshal event: %v", err)
				return errorx.Unknown
			}

			if err := wsClient.Write(b, true); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot write to ws: %v", err)
				return errorx.Unknown
			}
		}
	}
}
