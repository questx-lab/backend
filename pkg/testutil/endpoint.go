package testutil

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/api/twitter"
)

type MockTwitterEndpoint struct {
	GetUserFunc        func(context.Context, string) (twitter.User, error)
	GetTweetFunc       func(context.Context, string) (twitter.Tweet, error)
	CheckFollowingFunc func(context.Context, string) (bool, error)
	GetLikedTweetFunc  func(context.Context) ([]twitter.Tweet, error)
	GetRetweetFunc     func(context.Context, string) ([]twitter.Tweet, error)
}

func (e *MockTwitterEndpoint) WithUser(id string) {
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

type MockDiscordEndpoint struct {
	GetMeFunc       func(ctx context.Context, token string) (discord.User, error)
	HasAddedBotFunc func(ctx context.Context, guildID string) (bool, error)
	CheckMemberFunc func(ctx context.Context, guildID string) (bool, error)
	CheckCodeFunc   func(ctx context.Context, guildID, code string) error
	GetGuildFunc    func(ctx context.Context, guildID string) (discord.Guild, error)
	GetRolesFunc    func(ctx context.Context, guildID string) ([]discord.Role, error)
	GiveRoleFunc    func(ctx context.Context, guildID, roleID string) error
}

func (e *MockDiscordEndpoint) WithUser(id string) {
}

func (e *MockDiscordEndpoint) GetMe(ctx context.Context, token string) (discord.User, error) {
	if e.GetMeFunc != nil {
		return e.GetMeFunc(ctx, token)
	}

	return discord.User{}, errors.New("not implemented")
}

func (e *MockDiscordEndpoint) HasAddedBot(ctx context.Context, guildID string) (bool, error) {
	if e.HasAddedBotFunc != nil {
		return e.HasAddedBotFunc(ctx, guildID)
	}

	return false, errors.New("not implemented")
}

func (e *MockDiscordEndpoint) CheckMember(ctx context.Context, guildID string) (bool, error) {
	if e.CheckMemberFunc != nil {
		return e.CheckMemberFunc(ctx, guildID)
	}

	return false, errors.New("not implemented")
}

func (e *MockDiscordEndpoint) CheckCode(ctx context.Context, guildID, code string) error {
	if e.CheckCodeFunc != nil {
		return e.CheckCodeFunc(ctx, guildID, code)
	}

	return errors.New("not implemented")
}

func (e *MockDiscordEndpoint) GetGuild(ctx context.Context, guildID string) (discord.Guild, error) {
	if e.GetGuildFunc != nil {
		return e.GetGuildFunc(ctx, guildID)
	}

	return discord.Guild{}, errors.New("not implemented")
}

func (e *MockDiscordEndpoint) GetRoles(ctx context.Context, guildID string) ([]discord.Role, error) {
	if e.GetRolesFunc != nil {
		return e.GetRolesFunc(ctx, guildID)
	}

	return nil, errors.New("not implemented")
}

func (e *MockDiscordEndpoint) GiveRole(ctx context.Context, guildID, roleID string) error {
	if e.GetRolesFunc != nil {
		return e.GiveRoleFunc(ctx, guildID, roleID)
	}

	return errors.New("not implemented")
}
