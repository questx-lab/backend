package engine

import (
	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/domain/notification/event"
)

type ProxySession struct {
	C chan *event.EventRequest

	id         string
	processors map[string]*Processor
}

func NewProxySession() *ProxySession {
	return &ProxySession{
		C:          make(chan *event.EventRequest, 16),
		id:         uuid.NewString(),
		processors: make(map[string]*Processor),
	}
}

func (s *ProxySession) JoinProcessor(processor *Processor) {
	processor.Register(s)
	s.processors[processor.communityID] = processor
}

func (s *ProxySession) LeaveProcessor(processor *Processor) {
	processor.Unregister(s)
	delete(s.processors, processor.communityID)
}

func (s *ProxySession) LeaveAllProcessors() {
	for _, processor := range s.processors {
		processor.Unregister(s)
		delete(s.processors, processor.communityID)
	}
	close(s.C)
}
