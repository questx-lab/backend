package twitter

import "context"

type IEndpoint interface {
	GetUser(ctx context.Context, userScreenName string) (User, error)
	GetTweet(ctx context.Context, author, tweetID string) (Tweet, error)
	CheckAndGetReply(ctx context.Context, author, tweetID, replyTo string) (Tweet, error)
}
