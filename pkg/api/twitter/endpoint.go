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

	if user.Name == "" || user.Handle == "" {
		return User{}, errors.New("cannot get user info")
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
