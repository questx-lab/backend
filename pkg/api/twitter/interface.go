package twitter

import "context"

type IEndpoint interface {
	WithUser(string)
	OnBehalf() string
	GetUser(context.Context, string) (User, error)
	GetTweet(context.Context, string) (Tweet, error)
	GetLikedTweet(context.Context) ([]Tweet, error)
	GetRetweet(context.Context, string) ([]Tweet, error)
	CheckFollowing(context.Context, string) (bool, error)
}
