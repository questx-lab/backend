package twitter

import "context"

type IEndpoint interface {
	GetUser(ctx context.Context, userScreenName string) (User, error)
	GetTweet(ctx context.Context, author, tweetID string) (Tweet, error)
	GetReplyAndRetweet(ctx context.Context, handle, toAuthor, toTweetID string) (*Tweet, *Tweet, error)
}
