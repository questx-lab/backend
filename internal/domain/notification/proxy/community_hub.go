package proxy

import (
	"fmt"
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

		start := time.Now()
		h.mutex.RLock()
		start1 := time.Now()
		for _, s := range h.userSessions {
			s.C <- event
		}
		fmt.Println("BROADCAST 1 ELASED:", time.Since(start1))
		h.mutex.RUnlock()
		fmt.Println("BROADCAST ELASED:", time.Since(start))
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
