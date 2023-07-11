package client

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/questx-lab/backend/internal/domain/search"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type SearchCaller interface {
	IndexCommunity(ctx context.Context, id string, data search.CommunityData) error
	IndexQuest(ctx context.Context, id string, data search.QuestData) error
	DeleteCommunity(ctx context.Context, id string) error
	DeleteQuest(ctx context.Context, id string) error
	SearchCommunity(ctx context.Context, query string) ([]string, error)
	SearchQuest(ctx context.Context, query string, offset, limit int) ([]string, error)
	Close()
}

type searchCaller struct {
	client *rpc.Client
}

func NewSearchCaller(client *rpc.Client) *searchCaller {
	return &searchCaller{client: client}
}

func (c *searchCaller) IndexCommunity(ctx context.Context, id string, data search.CommunityData) error {
	return c.client.CallContext(ctx, nil, c.fname(ctx, "index"), search.CommunityDoc, id, data)
}

func (c *searchCaller) IndexQuest(ctx context.Context, id string, data search.QuestData) error {
	return c.client.CallContext(ctx, nil, c.fname(ctx, "index"), search.QuestDoc, id, data)
}

func (c *searchCaller) DeleteCommunity(ctx context.Context, id string) error {
	return c.client.CallContext(ctx, nil, c.fname(ctx, "delete"), search.CommunityDoc, id)
}

func (c *searchCaller) DeleteQuest(ctx context.Context, id string) error {
	return c.client.CallContext(ctx, nil, c.fname(ctx, "delete"), search.QuestDoc, id)
}

func (c *searchCaller) SearchCommunity(ctx context.Context, query string) ([]string, error) {
	var result []string
	err := c.client.CallContext(ctx, &result, c.fname(ctx, "search"), search.CommunityDoc, query, 0, int(1e6))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *searchCaller) SearchQuest(ctx context.Context, query string, offset, limit int) ([]string, error) {
	var result []string
	err := c.client.
		CallContext(ctx, &result, c.fname(ctx, "search"), search.QuestDoc, query, offset, limit)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *searchCaller) Close() {
	c.client.Close()
}

func (c *searchCaller) fname(ctx context.Context, funcName string) string {
	return fmt.Sprintf("%s_%s", xcontext.Configs(ctx).SearchServer.RPCName, funcName)
}
