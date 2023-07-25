package proxy

import (
	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/domain/notification/event"
)

type UserSession struct {
	C chan *event.EventRequest

	id                  string
	userID              string
	joinedCommunityHubs []*CommunityHub
	joinedUserHub       *UserHub
}

func NewUserSession(userID string) *UserSession {
	session := &UserSession{
		C:                   make(chan *event.EventRequest, 16),
		id:                  uuid.NewString(),
		userID:              userID,
		joinedCommunityHubs: make([]*CommunityHub, 0),
		joinedUserHub:       nil,
	}

	return session
}

func (s *UserSession) JoinCommunity(hub *CommunityHub) {
	hub.register(s)
	s.joinedCommunityHubs = append(s.joinedCommunityHubs, hub)
}

func (s *UserSession) JoinUser(hub *UserHub) {
	hub.register(s)
	s.joinedUserHub = hub
}

func (s *UserSession) Leave() {
	for _, hub := range s.joinedCommunityHubs {
		hub.unregister(s)
	}

	s.joinedCommunityHubs = s.joinedCommunityHubs[:0]
	s.joinedUserHub.unregister(s)

	close(s.C)
}
