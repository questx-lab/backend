package proxy

import (
	"sync"

	"github.com/questx-lab/backend/internal/domain/notification/event"
)

type UserHub struct {
	userID       string
	userSessions map[string]*UserSession

	mutex sync.RWMutex
}

func NewUserHub(userID string) *UserHub {
	h := &UserHub{
		userID:       userID,
		userSessions: make(map[string]*UserSession),
		mutex:        sync.RWMutex{},
	}

	return h
}

func (h *UserHub) Send(event *event.EventRequest) {
	h.mutex.RLock()
	for _, s := range h.userSessions {
		s.C <- event
	}
	h.mutex.RUnlock()
}

func (h *UserHub) register(session *UserSession) {
	h.mutex.RLock()
	_, ok := h.userSessions[session.id]
	h.mutex.RUnlock()
	if ok {
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Double check.
	if _, ok := h.userSessions[session.id]; !ok {
		h.userSessions[session.id] = session
	}
}

func (h *UserHub) unregister(session *UserSession) {
	h.mutex.RLock()
	_, ok := h.userSessions[session.id]
	h.mutex.RUnlock()
	if !ok {
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()
	delete(h.userSessions, session.id)
}

func (h *UserHub) IsEmpty() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return len(h.userSessions) == 0
}
