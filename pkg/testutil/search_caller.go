package testutil

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/domain/search"
)

type MockSearchCaller struct {
	IndexCommunityFunc   func(ctx context.Context, id string, data search.CommunityData) error
	IndexQuestFunc       func(ctx context.Context, id string, data search.QuestData) error
	DeleteCommunityFunc  func(ctx context.Context, id string) error
	DeleteQuestFunc      func(ctx context.Context, id string) error
	ReplaceCommunityFunc func(ctx context.Context, id string, data search.CommunityData) error
	ReplaceQuestFunc     func(ctx context.Context, id string, data search.QuestData) error
	SearchCommunityFunc  func(ctx context.Context, query string, offset, limit int) ([]string, error)
	SearchQuestFunc      func(ctx context.Context, query string, offset, limit int) ([]string, error)
}

func (c *MockSearchCaller) IndexCommunity(ctx context.Context, id string, data search.CommunityData) error {
	if c.IndexCommunityFunc != nil {
		return c.IndexCommunityFunc(ctx, id, data)
	}

	return nil
}

func (c *MockSearchCaller) IndexQuest(ctx context.Context, id string, data search.QuestData) error {
	if c.IndexQuestFunc != nil {
		return c.IndexQuestFunc(ctx, id, data)
	}

	return nil
}

func (c *MockSearchCaller) DeleteCommunity(ctx context.Context, id string) error {
	if c.DeleteCommunityFunc != nil {
		return c.DeleteCommunityFunc(ctx, id)
	}

	return nil
}

func (c *MockSearchCaller) DeleteQuest(ctx context.Context, id string) error {
	if c.DeleteQuestFunc != nil {
		return c.DeleteQuestFunc(ctx, id)
	}

	return nil
}

func (c *MockSearchCaller) ReplaceCommunity(ctx context.Context, id string, data search.CommunityData) error {
	if c.ReplaceCommunityFunc != nil {
		return c.ReplaceCommunityFunc(ctx, id, data)
	}

	return nil
}

func (c *MockSearchCaller) ReplaceQuest(ctx context.Context, id string, data search.QuestData) error {
	if c.ReplaceQuestFunc != nil {
		return c.ReplaceQuestFunc(ctx, id, data)
	}

	return nil
}

func (c *MockSearchCaller) SearchCommunity(ctx context.Context, query string, offset, limit int) ([]string, error) {
	if c.SearchCommunityFunc != nil {
		return c.SearchCommunityFunc(ctx, query, offset, limit)
	}

	return nil, errors.New("not implemented")
}

func (c *MockSearchCaller) SearchQuest(ctx context.Context, query string, offset, limit int) ([]string, error) {
	if c.SearchQuestFunc != nil {
		return c.SearchQuestFunc(ctx, query, offset, limit)
	}

	return nil, errors.New("not implemented")
}
