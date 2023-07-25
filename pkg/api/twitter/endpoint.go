package twitter

import (
	"context"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/api"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type Endpoint struct {
	apiGenerator api.Generator
}

func New(cfg config.TwitterConfigs) *Endpoint {
	return &Endpoint{
		apiGenerator: api.NewGenerator(cfg.APIEndpoints...),
	}
}

func (e *Endpoint) GetUser(ctx context.Context, userScreenName string) (User, error) {
	resp, err := e.apiGenerator.New("/get_user").
		Query(api.Parameter{"handle": userScreenName}).
		GET(ctx)
	if err != nil {
		return User{}, err
	}

	if resp.Code != 200 {
		xcontext.Logger(ctx).Errorf("Invalid status code: %v", resp.Body)
		return User{}, fmt.Errorf("invalid status code %d", resp.Code)
	}

	body, ok := resp.Body.(api.JSON)
	if !ok {
		return User{}, errors.New("invalid body format")
	}

	user := User{}
	if err = mapstructure.Decode(body, &user); err != nil {
		return User{}, nil
	}

	return user, nil
}

func (e *Endpoint) GetTweet(ctx context.Context, author string, tweetID string) (Tweet, error) {
	resp, err := e.apiGenerator.New("/get_tweet").
		Query(api.Parameter{
			"author":   author,
			"tweet_id": tweetID,
		}).
		GET(ctx)

	if err != nil {
		return Tweet{}, err
	}

	if resp.Code != 200 {
		xcontext.Logger(ctx).Errorf("Invalid status code: %v", resp.Body)
		return Tweet{}, fmt.Errorf("invalid status code %d", resp.Code)
	}

	body, ok := resp.Body.(api.JSON)
	if !ok {
		return Tweet{}, errors.New("invalid body format")
	}

	tweet := Tweet{}
	if err := mapstructure.Decode(body, &tweet); err != nil {
		return Tweet{}, nil
	}

	return tweet, nil
}

func (e *Endpoint) CheckFollowing(ctx context.Context, source, target string) (bool, error) {
	resp, err := e.apiGenerator.New("/is_user_following").
		Query(api.Parameter{
			"source": source,
			"target": target,
		}).
		GET(ctx)
	if err != nil {
		return false, err
	}

	if resp.Code != 200 {
		xcontext.Logger(ctx).Errorf("Invalid status code: %v", resp.Body)
		return false, fmt.Errorf("invalid status code %d", resp.Code)
	}

	body, ok := resp.Body.(api.JSON)
	if !ok {
		return false, errors.New("invalid resp")
	}

	return body.GetBool("result")
}

func (e *Endpoint) CheckLiked(ctx context.Context, handle, toAuthor, toTweetID string) (bool, error) {
	resp, err := e.apiGenerator.New("/is_user_liked").
		Query(api.Parameter{
			"handle":      handle,
			"to_author":   toAuthor,
			"to_tweet_id": toTweetID,
		}).
		GET(ctx)
	if err != nil {
		return false, err
	}

	if resp.Code != 200 {
		xcontext.Logger(ctx).Errorf("Invalid status code: %v", resp.Body)
		return false, fmt.Errorf("invalid status code %d", resp.Code)
	}

	body, ok := resp.Body.(api.JSON)
	if !ok {
		return false, errors.New("invalid body format")
	}

	return body.GetBool("result")
}

func (e *Endpoint) GetReplyAndRetweet(ctx context.Context, handle, toAuthor, toTweetID string) (*Tweet, *Tweet, error) {
	resp, err := e.apiGenerator.New("/get_reply_and_retweet").
		Query(api.Parameter{
			"handle":      handle,
			"to_author":   toAuthor,
			"to_tweet_id": toTweetID,
		}).
		GET(ctx)

	if err != nil {
		return nil, nil, err
	}

	body, ok := resp.Body.(api.JSON)
	if !ok {
		return nil, nil, errors.New("invalid body format")
	}

	var reply *Tweet
	var retweet *Tweet

	replyMap, ok := body["reply"]
	if ok {
		replyTmp := Tweet{}
		if err := mapstructure.Decode(replyMap, &replyTmp); err != nil {
			return nil, nil, err
		}

		reply = &replyTmp
	}

	retweetMap, ok := body["retweet"]
	if ok {
		retweetTmp := Tweet{}
		if err := mapstructure.Decode(retweetMap, &retweetTmp); err != nil {
			return nil, nil, err
		}

		retweet = &retweetTmp
	}

	return reply, retweet, nil
}
