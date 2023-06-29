package twitter

import "context"

type IEndpoint interface {
	GetUser(ctx context.Context, userScreenName string) (User, error)
	GetTweet(ctx context.Context, author, tweetID string) (Tweet, error)
	CheckLiked(ctx context.Context, handle, toAuthor, toTweetID string) (bool, error)
	GetReplyAndRetweet(ctx context.Context, handle, toAuthor, toTweetID string) (*Tweet, *Tweet, error)
	CheckFollowing(ctx context.Context, source, target string) (bool, error)
}
