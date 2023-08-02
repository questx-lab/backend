package proxy

import (
	"log"
	"sync"
	"time"

	"github.com/questx-lab/backend/internal/domain/notification/event"
)

type CommunityHub struct {
	communityID  string
	userSessions map[string]*UserSession
	c            chan *event.EventRequest

	mutex sync.RWMutex
}

func NewCommunityHub(communityID string) *CommunityHub {
	h := &CommunityHub{
		communityID:  communityID,
		userSessions: make(map[string]*UserSession),
		mutex:        sync.RWMutex{},
		c:            make(chan *event.EventRequest, 8),
	}

	go h.run()
	return h
}

func (h *CommunityHub) run() {
	for {
		event, ok := <-h.c
		if !ok {
			break
		}

		start0 := time.Now()
		h.mutex.RLock()
		start := time.Now()
		for _, s := range h.userSessions {
			s.C <- event
		}
		log.Println("BROADCAST", time.Since(start))
		h.mutex.RUnlock()
		log.Println("BROADCAST1", time.Since(start0))
	}
}

func (h *CommunityHub) register(session *UserSession) {
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

func (h *CommunityHub) unregister(session *UserSession) {
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

func (h *CommunityHub) IsEmpty() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return len(h.userSessions) == 0
}
