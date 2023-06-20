package pubsub

import (
	"context"
	"time"
)

type SubscribeHandler func(ctx context.Context, topic string, pack *Pack, tt time.Time)

type Subscriber interface {
	Subscribe(context.Context)
	Stop(ctx context.Context) error
}
