package engine

import (
	"sync"

	"github.com/questx-lab/backend/internal/domain/notification/event"
)

type Processor struct {
	communityID string
	proxies     map[string]*ProxySession
	mutex       sync.RWMutex
}

func NewProcessor(communityID string) *Processor {
	return &Processor{
		communityID: communityID,
		proxies:     make(map[string]*ProxySession),
		mutex:       sync.RWMutex{},
	}
}

func (p *Processor) Register(session *ProxySession) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, ok := p.proxies[session.id]; ok {
		return
	}

	p.proxies[session.id] = session
}

func (p *Processor) Unregister(session *ProxySession) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	delete(p.proxies, session.id)
}

func (p *Processor) Broadcast(ev *event.EventRequest) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, proxy := range p.proxies {
		proxy.C <- ev
	}
}
