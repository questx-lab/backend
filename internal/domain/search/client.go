package search

import (
	"context"
	"fmt"

	"github.com/questx-lab/backend/pkg/xcontext"
)

const (
	communityDoc = "community"
	questDoc     = "quest"
	templateDoc  = "template"
)

type CommunityData struct {
	Name         string
	Introduction string
}

type QuestData struct {
	Title       string
	Description string
}

type TemplateData struct {
	Title       string
	Description string
}

type Caller interface {
	IndexCommunity(ctx context.Context, id string, data CommunityData) error
	IndexQuest(ctx context.Context, id string, data QuestData) error
	DeleteCommunity(ctx context.Context, id string) error
	DeleteQuest(ctx context.Context, id string) error
	ReplaceCommunity(ctx context.Context, id string, data CommunityData) error
	ReplaceQuest(ctx context.Context, id string, data QuestData) error
	SearchCommunity(ctx context.Context, query string, offset, limit int) ([]string, error)
	SearchQuest(ctx context.Context, query string, offset, limit int) ([]string, error)
}

type caller struct{}

func NewCaller() *caller {
	return &caller{}
}

func (c *caller) IndexCommunity(ctx context.Context, id string, data CommunityData) error {
	return xcontext.RPCSearchClient(ctx).
		CallContext(ctx, nil, c.rpcFuncName(ctx, "index"), communityDoc, id, data)
}

func (c *caller) IndexQuest(ctx context.Context, id string, data QuestData) error {
	return xcontext.RPCSearchClient(ctx).
		CallContext(ctx, nil, c.rpcFuncName(ctx, "index"), questDoc, id, data)
}

func (c *caller) DeleteCommunity(ctx context.Context, id string) error {
	return xcontext.RPCSearchClient(ctx).
		CallContext(ctx, nil, c.rpcFuncName(ctx, "delete"), communityDoc, id)
}

func (c *caller) DeleteQuest(ctx context.Context, id string) error {
	return xcontext.RPCSearchClient(ctx).
		CallContext(ctx, nil, c.rpcFuncName(ctx, "delete"), questDoc, id)
}

func (c *caller) ReplaceCommunity(ctx context.Context, id string, data CommunityData) error {
	return xcontext.RPCSearchClient(ctx).
		CallContext(ctx, nil, c.rpcFuncName(ctx, "replace"), communityDoc, id, data)
}

func (c *caller) ReplaceQuest(ctx context.Context, id string, data QuestData) error {
	return xcontext.RPCSearchClient(ctx).
		CallContext(ctx, nil, c.rpcFuncName(ctx, "replace"), questDoc, id, data)
}

func (c *caller) SearchCommunity(ctx context.Context, query string, offset, limit int) ([]string, error) {
	var result []string
	err := xcontext.RPCSearchClient(ctx).
		CallContext(ctx, &result, c.rpcFuncName(ctx, "search"), communityDoc, query, offset, limit)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *caller) SearchQuest(ctx context.Context, query string, offset, limit int) ([]string, error) {
	var result []string
	err := xcontext.RPCSearchClient(ctx).
		CallContext(ctx, &result, c.rpcFuncName(ctx, "search"), questDoc, query, offset, limit)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *caller) rpcFuncName(ctx context.Context, funcName string) string {
	return fmt.Sprintf("%s_%s", xcontext.Configs(ctx).SearchServer.RPCName, funcName)
}
