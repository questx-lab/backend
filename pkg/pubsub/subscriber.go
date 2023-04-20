package pubsub

import (
	"context"
	"time"
)

type SubscribeHandler func(context.Context, *Pack, time.Time)

type Subscriber interface {
	Subscribe(context.Context)
	Stop(ctx context.Context) error
}
