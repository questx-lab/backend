package client

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type GameEngineCaller interface {
	StartRoom(ctx context.Context, roomID string) error
	StopRoom(ctx context.Context, roomID string) error
	StartLuckyboxEvent(ctx context.Context, eventID, roomID string) error
	StopLuckyboxEvent(ctx context.Context, eventID, roomID string) error
	Close()
}

type gameEngineCaller struct {
	client *rpc.Client
}

func NewGameEngineCaller(client *rpc.Client) *gameEngineCaller {
	return &gameEngineCaller{client: client}
}

func (c *gameEngineCaller) StartRoom(ctx context.Context, roomID string) error {
	return c.client.CallContext(ctx, nil, c.fname(ctx, "startRoom"), roomID)
}

func (c *gameEngineCaller) StopRoom(ctx context.Context, roomID string) error {
	return c.client.CallContext(ctx, nil, c.fname(ctx, "stopRoom"), roomID)
}

func (c *gameEngineCaller) StartLuckyboxEvent(ctx context.Context, eventID, roomID string) error {
	return c.client.CallContext(ctx, nil, c.fname(ctx, "startLuckyboxEvent"), eventID, roomID)
}

func (c *gameEngineCaller) StopLuckyboxEvent(ctx context.Context, eventID, roomID string) error {
	return c.client.CallContext(ctx, nil, c.fname(ctx, "stopLuckyboxEvent"), eventID, roomID)
}

func (c *gameEngineCaller) Close() {
	c.client.Close()
}

func (c *gameEngineCaller) fname(ctx context.Context, funcName string) string {
	return fmt.Sprintf("%s_%s", xcontext.Configs(ctx).GameEngineRPCServer.RPCName, funcName)
}
