package testutil

import (
	"context"

	"github.com/questx-lab/backend/pkg/pubsub"
)

type MockPublisher struct {
	PublishFunc func(context.Context, string, *pubsub.Pack) error
}

func (p *MockPublisher) Publish(ctx context.Context, topic string, pack *pubsub.Pack) error {
	if p.PublishFunc != nil {
		return p.PublishFunc(ctx, topic, pack)
	}

	return nil
}
