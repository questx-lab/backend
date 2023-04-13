package pubsub

import (
	"context"
	"time"
)

type SubscribeHandler func(context.Context, *Pack, time.Time)

type Subscriber interface {
	Subscribe(context.Context, *Pack, time.Time)
	Stop(ctx context.Context) error
}
