package twitter

import "context"

type IEndpoint interface {
	GetUser(ctx context.Context, userScreenName string) (User, error)
	GetTweet(ctx context.Context, tweetID string) (Tweet, error)
	GetLikedTweet(ctx context.Context, userScreenName string) ([]Tweet, error)
	GetRetweet(ctx context.Context, tweetID string) ([]Tweet, error)
	CheckFollowing(ctx context.Context, source, target string) (bool, error)
}
