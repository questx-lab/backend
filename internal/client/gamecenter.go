package client

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type GameCenterCaller interface {
	Ping(ctx context.Context, domainName string, isNew bool) error
	StartRoom(ctx context.Context, roomID string) error
	CreateCharacter(ctx context.Context, characterID string) error
	BuyCharacter(ctx context.Context, userID, characterID, communityID string) error
	Close()
}

type gameCenterCaller struct {
	client *rpc.Client
}

func NewGameCenterCaller(client *rpc.Client) *gameCenterCaller {
	return &gameCenterCaller{client: client}
}

func (c *gameCenterCaller) Ping(ctx context.Context, domainName string, isNew bool) error {
	return c.client.CallContext(ctx, nil, c.fname(ctx, "ping"), domainName, isNew)
}

func (c *gameCenterCaller) StartRoom(ctx context.Context, roomID string) error {
	return c.client.CallContext(ctx, nil, c.fname(ctx, "startRoom"), roomID)
}

func (c *gameCenterCaller) CreateCharacter(ctx context.Context, characterID string) error {
	return c.client.CallContext(ctx, nil, c.fname(ctx, "createCharacter"), characterID)
}

func (c *gameCenterCaller) BuyCharacter(ctx context.Context, userID, characterID, communityID string) error {
	return c.client.CallContext(ctx, nil, c.fname(ctx, "buyCharacter"), userID, characterID, communityID)
}

func (c *gameCenterCaller) Close() {
	c.client.Close()
}

func (c *gameCenterCaller) fname(ctx context.Context, funcName string) string {
	return fmt.Sprintf("%s_%s", xcontext.Configs(ctx).GameCenterServer.RPCName, funcName)
}
