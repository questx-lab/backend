package client

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/internal/domain/notification/event"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type NotificationEngineCaller interface {
	Emit(ctx context.Context, ev *event.EventRequest) error
	Close()
}

type notificationEngineCaller struct {
	client *rpc.Client
}

func NewNotificationEngineCaller(client *rpc.Client) *notificationEngineCaller {
	return &notificationEngineCaller{client: client}
}

func (c *notificationEngineCaller) Emit(ctx context.Context, ev *event.EventRequest) error {
	return c.client.CallContext(ctx, nil, c.fname(ctx, "emit"), ev)
}

func (c *notificationEngineCaller) Close() {
	c.client.Close()
}

func (c *notificationEngineCaller) fname(ctx context.Context, funcName string) string {
	return fmt.Sprintf("%s_%s", xcontext.Configs(ctx).Notification.EngineRPCServer.RPCName, funcName)
}
