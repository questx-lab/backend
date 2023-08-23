package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/questx-lab/backend/internal/domain/notification/directive"
	"github.com/questx-lab/backend/internal/domain/notification/event"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"google.golang.org/api/option"
)

type EngineServer struct {
	rootCtx context.Context

	messagingClient    *messaging.Client
	communityProcessor map[string]*CommunityProcessor
	userProcessors     map[string]*UserProcessor
	mutex              sync.RWMutex
	seq                uint64
}

func NewEngineServer(ctx context.Context) *EngineServer {
	opt := option.WithCredentialsJSON([]byte(xcontext.Configs(ctx).Auth.Google.AuthenticationCredentialsJSON))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic(err)
	}

	messagingClient, err := app.Messaging(ctx)
	if err != nil {
		panic(err)
	}

	return &EngineServer{
		rootCtx:            ctx,
		messagingClient:    messagingClient,
		communityProcessor: make(map[string]*CommunityProcessor),
		userProcessors:     make(map[string]*UserProcessor),
		mutex:              sync.RWMutex{},
		seq:                0,
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
func (s *EngineServer) Emit(_ context.Context, ev *event.EventRequest) error {
	seq := atomic.AddUint64(&s.seq, 1) - 1
	ev.Seq = seq

	var message *messaging.Message
	data, origin, err := event.FormatMessaging(ev)
	if err != nil {
		xcontext.Logger(s.rootCtx).Errorf("Cannot format message: %v", err)
	} else {
		message = &messaging.Message{
			Data:       data,
			FCMOptions: &messaging.FCMOptions{AnalyticsLabel: origin.Op()},
		}
	}

	for _, community := range ev.Metadata.ToCommunities {
		processor := s.GetCommunityProcessor(community.ID, false)
		if processor != nil {
			processor.Broadcast(ev)
		}

		if message != nil {
			message.Topic = fmt.Sprintf("community~%s", community.ID)

			switch origin.Op() {
			case event.MessageCreatedEvent{}.Op():
				messageEvent, ok := origin.(*event.MessageCreatedEvent)
				if !ok {
					xcontext.Logger(s.rootCtx).Errorf("Cannot parse raw event to message")
					break
				}

				message.Notification = &messaging.Notification{
					Title: fmt.Sprintf("A new chat from %s", community.Handle),
					Body:  messageEvent.Content,
				}
			}

			if _, err := s.messagingClient.Send(s.rootCtx, message); err != nil {
				xcontext.Logger(s.rootCtx).Errorf("Cannot send message to community: %v", err)
			}
		}
	}

	for _, userID := range ev.Metadata.ToUsers {
		processor := s.GetUserProcessor(userID, false)
		if processor != nil {
			processor.Send(ev)
		}

		if message != nil {
			message.Topic = fmt.Sprintf("user~%s", userID)
			if _, err := s.messagingClient.Send(s.rootCtx, message); err != nil {
				xcontext.Logger(s.rootCtx).Errorf("Cannot send message to user: %v", err)
			}
		}
	}
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
