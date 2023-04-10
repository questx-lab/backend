package testutil

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/pkg/api/twitter"
)

type MockTwitterEndpoint struct {
	GetUserFunc        func(context.Context, string) (twitter.User, error)
	GetTweetFunc       func(context.Context, string) (twitter.Tweet, error)
	CheckFollowingFunc func(context.Context, string) (bool, error)
	GetLikedTweetFunc  func(context.Context) ([]twitter.Tweet, error)
	GetRetweetFunc     func(context.Context, string) ([]twitter.Tweet, error)
}

func (e *MockTwitterEndpoint) WithUser(id string) twitter.IEndpoint {
	return e
}

func (e *MockTwitterEndpoint) OnBehalf() string {
	return "foo"
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

func (e *MockTwitterEndpoint) GetLikedTweet(ctx context.Context) ([]twitter.Tweet, error) {
	if e.GetLikedTweetFunc != nil {
		return e.GetLikedTweetFunc(ctx)
	}

	return nil, errors.New("not implemented")
}

func (e *MockTwitterEndpoint) GetRetweet(ctx context.Context, tweetID string) ([]twitter.Tweet, error) {
	if e.GetRetweetFunc != nil {
		return e.GetRetweetFunc(ctx, tweetID)
	}

	return nil, errors.New("not implemented")
}

func (e *MockTwitterEndpoint) GetTweet(ctx context.Context, tweetID string) (twitter.Tweet, error) {
	if e.GetTweetFunc != nil {
		return e.GetTweetFunc(ctx, tweetID)
	}

	return twitter.Tweet{}, errors.New("not implemented")
}
