package proxy

import (
	"fmt"
	"sync"
	"time"

	"github.com/questx-lab/backend/internal/domain/notification/event"
)

type Hub struct {
	communityID string
	sessions    map[string]*Session
	c           chan *event.EventRequest

	mutex sync.RWMutex
}

func NewHub(communityID string) *Hub {
	h := &Hub{
		communityID: communityID,
		sessions:    make(map[string]*Session),
		mutex:       sync.RWMutex{},
		c:           make(chan *event.EventRequest, 8),
	}

	go h.run()
	return h
}

func (h *Hub) run() {
	for {
		event, ok := <-h.c
		if !ok {
			break
		}

		start := time.Now()
		h.mutex.RLock()
		start1 := time.Now()
		for _, s := range h.sessions {
			s.C <- event
		}
		fmt.Println("BROADCAST 1 ELASED:", time.Since(start1))
		h.mutex.RUnlock()
		fmt.Println("BROADCAST ELASED:", time.Since(start))
	}
}

func (h *Hub) Register(session *Session) {
	h.mutex.RLock()
	_, ok := h.sessions[session.id]
	h.mutex.RUnlock()
	if ok {
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Double check.
	if _, ok := h.sessions[session.id]; !ok {
		h.sessions[session.id] = session
	}
}

func (h *Hub) Unregister(session *Session) {
	h.mutex.RLock()
	_, ok := h.sessions[session.id]
	h.mutex.RUnlock()
	if !ok {
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()
	delete(h.sessions, session.id)
}

func (h *Hub) IsEmpty() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return len(h.sessions) == 0
}
