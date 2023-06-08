package testutil

import (
	"context"

	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/pubsub"
)

type MockPublisher struct {
	PublishFunc func(context.Context, string, *pubsub.Pack) error
}

func (m *MockPublisher) Publish(ctx context.Context, topic string, pack *pubsub.Pack) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, topic, pack)
	}

	return errorx.New(errorx.NotImplemented, "Not implemented")
}
