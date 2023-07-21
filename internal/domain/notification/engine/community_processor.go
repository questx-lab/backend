package engine

import (
	"sync"

	"github.com/questx-lab/backend/internal/domain/notification/event"
)

type CommunityProcessor struct {
	communityID string
	proxies     map[string]*ProxySession
	mutex       sync.RWMutex
}

func NewCommunityProcessor(communityID string) *CommunityProcessor {
	return &CommunityProcessor{
		communityID: communityID,
		proxies:     make(map[string]*ProxySession),
		mutex:       sync.RWMutex{},
	}
}

func (p *CommunityProcessor) register(proxy *ProxySession) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, ok := p.proxies[proxy.id]; ok {
		return
	}

	p.proxies[proxy.id] = proxy
}

func (p *CommunityProcessor) unregister(proxy *ProxySession) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delete(p.proxies, proxy.id)
}

func (p *CommunityProcessor) Broadcast(ev *event.EventRequest) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for _, proxy := range p.proxies {
		proxy.C <- ev
	}
}
