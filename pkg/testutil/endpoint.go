package testutil

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/pkg/api/twitter"
)

type MockTwitterEndpoint struct {
	GetUserFunc        func(context.Context, string) (twitter.User, error)
	CheckFollowingFunc func(context.Context, string) (bool, error)
}

func (e *MockTwitterEndpoint) WithUser(id string) twitter.IEndpoint {
	return e
}

func (e *MockTwitterEndpoint) GetUser(ctx context.Context, id string) (twitter.User, error) {
	if e.GetUserFunc != nil {
		return e.GetUserFunc(ctx, id)
	}

	return twitter.User{}, errors.New("not implemented")
}

func (e *MockTwitterEndpoint) CheckFollowing(ctx context.Context, id string) (bool, error) {
	if e.CheckFollowingFunc != nil {
		return e.CheckFollowingFunc(ctx, id)
	}

	return false, errors.New("not implemented")
}
