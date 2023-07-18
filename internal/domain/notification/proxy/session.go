package proxy

import (
	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/domain/notification/event"
)

type Session struct {
	C chan *event.EventRequest

	id         string
	joinedHubs []*Hub
}

func NewSession() *Session {
	session := &Session{
		C:          make(chan *event.EventRequest, 16),
		id:         uuid.NewString(),
		joinedHubs: make([]*Hub, 0),
	}

	return session
}

func (s *Session) JoinHub(hub *Hub) {
	hub.Register(s)
	s.joinedHubs = append(s.joinedHubs, hub)
}

func (s *Session) LeaveAllHubs() {
	for _, hub := range s.joinedHubs {
		hub.Unregister(s)
	}
	close(s.C)
}
