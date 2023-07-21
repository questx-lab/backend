package engine

import (
	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/domain/notification/event"
)

type ProxySession struct {
	C chan *event.EventRequest

	id                  string
	communityProcessors map[string]*CommunityProcessor
	userProcessor       map[string]*UserProcessor
}

func NewProxySession() *ProxySession {
	return &ProxySession{
		C:                   make(chan *event.EventRequest, 16),
		id:                  uuid.NewString(),
		communityProcessors: make(map[string]*CommunityProcessor),
		userProcessor:       make(map[string]*UserProcessor),
	}
}

func (s *ProxySession) RegisterCommunity(community *CommunityProcessor) {
	if community == nil {
		return
	}

	community.register(s)
	s.communityProcessors[community.communityID] = community
}

func (s *ProxySession) UnregisterCommunity(community *CommunityProcessor) {
	if community == nil {
		return
	}

	community.unregister(s)
	delete(s.communityProcessors, community.communityID)
}

func (s *ProxySession) RegisterUser(user *UserProcessor) {
	if user == nil {
		return
	}

	user.register(s)
	s.userProcessor[user.userID] = user
}

func (s *ProxySession) UnregisterUser(user *UserProcessor) {
	if user == nil {
		return
	}

	user.unregister(s)
	delete(s.userProcessor, user.userID)
}

func (s *ProxySession) Leave() {
	for _, community := range s.communityProcessors {
		s.UnregisterCommunity(community)
	}

	for _, user := range s.userProcessor {
		s.UnregisterUser(user)
	}

	close(s.C)
}
