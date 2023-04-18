package pubsub

import "context"

type Publisher interface {
	Publish(context.Context, string, *Pack) error
}
