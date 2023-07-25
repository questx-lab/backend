package engine

import (
	"sync"

	"github.com/questx-lab/backend/internal/domain/notification/event"
)

type UserProcessor struct {
	userID  string
	proxies map[string]*ProxySession
	mutex   sync.RWMutex
}

func NewUserProcessor(userID string) *UserProcessor {
	session := &UserProcessor{
		userID:  userID,
		proxies: make(map[string]*ProxySession),
		mutex:   sync.RWMutex{},
	}

	return session
}

func (s *UserProcessor) register(proxy *ProxySession) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.proxies[proxy.id]; ok {
		return
	}

	s.proxies[proxy.id] = proxy
}

func (s *UserProcessor) unregister(proxy *ProxySession) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.proxies, proxy.id)
}

func (s *UserProcessor) Send(ev *event.EventRequest) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, proxy := range s.proxies {
		proxy.C <- ev
	}
}
