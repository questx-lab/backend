package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/questx-lab/backend/internal/domain/notification/directive"
	"github.com/questx-lab/backend/internal/domain/notification/event"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type EngineServer struct {
	processors map[string]*Processor
	mutex      sync.RWMutex
}

func NewEngineServer() *EngineServer {
	return &EngineServer{
		processors: make(map[string]*Processor),
		mutex:      sync.RWMutex{},
	}
}

func (s *EngineServer) GetProcessor(communityID string) *Processor {
	s.mutex.RLock()
	processor, ok := s.processors[communityID]
	s.mutex.RUnlock()

	if ok {
		return processor
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Double check.
	if processor, ok := s.processors[communityID]; ok {
		return processor
	}

	s.processors[communityID] = NewProcessor(communityID)
	return s.processors[communityID]
}

// Emit handles a emit call from client. It broadcasts the event to every proxy
// registered to the community.
func (s *EngineServer) Emit(_ context.Context, event *event.EventRequest) error {
	start := time.Now()
	s.GetProcessor(event.Metadata.To).Broadcast(event)
	fmt.Println("EMIT ELAPSED: ", time.Since(start))
	return nil
}

func (s *EngineServer) ServeProxy(ctx context.Context, _ *model.ServeNotificationEngineRequest) error {
	session := NewProxySession()
	defer session.LeaveAllProcessors()

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
					xcontext.Logger(ctx).Errorf("Cannot unmarshal register directive data: %v", err)
					return errorx.Unknown
				}

				processor := s.GetProcessor(registerDirective.CommunityID)
				session.JoinProcessor(processor)

				xcontext.Logger(ctx).Infof("Proxy %s register to community %s",
					session.id, registerDirective.CommunityID)

			case directive.EngineUnregisterCommunityDirectiveOp:
				var unregisterDirective directive.EngineUnregisterCommunityDirective
				if err := json.Unmarshal(d.Data, &unregisterDirective); err != nil {
					xcontext.Logger(ctx).Errorf("Cannot unmarshal unregister directive data: %v", err)
					return errorx.Unknown
				}

				processor := s.GetProcessor(unregisterDirective.CommunityID)
				session.LeaveProcessor(processor)
				xcontext.Logger(ctx).Infof("Proxy %s unregister to community %s",
					session.id, unregisterDirective.CommunityID)
			}

		case ev, ok := <-session.C:
			if !ok {
				return errorx.Unknown
			}

			b, err := json.Marshal(ev)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot marshal event: %v", err)
				return errorx.Unknown
			}

			start := time.Now()
			if err := wsClient.Write(b, true); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot write to ws: %v", err)
				return errorx.Unknown
			}
			fmt.Println("WRITE ELAPSED: ", time.Since(start))
		}
	}
}
