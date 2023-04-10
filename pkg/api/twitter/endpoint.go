package twitter

import (
	"context"
	"errors"

	"github.com/mitchellh/mapstructure"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/api"
)

const ApiURL = "https://api.twitter.com"

var ErrRateLimit = errors.New("rate limit")

type Endpoint struct {
	// OAuth2.0 application only Access Token - for access to public api v2.
	AppToken string

	// OAuth1.0 developer Access Token - for access to public api v1.
	ConsumerKey string
	AccessToken string
	SigningKey  string

	// Twitter user id.
	UserID string
}

func New(ctx context.Context, cfg config.TwitterConfigs) (*Endpoint, error) {
	signingKey := api.PercentEncode(cfg.ConsumerAPISecret) +
		"&" + api.PercentEncode(cfg.AccessTokenSecret)

	return &Endpoint{
		AppToken:    cfg.AppAccessToken,
		ConsumerKey: cfg.ConsumerAPIKey,
		AccessToken: cfg.AccessToken,
		SigningKey:  signingKey,
	}, nil
}

func (e *Endpoint) WithUser(id string) IEndpoint {
	clone := *e
	clone.UserID = id
	return &clone
}

func (e *Endpoint) GetUser(ctx context.Context, userID string) (User, error) {
	resp, err := api.New(ApiURL, "/1.1/users/show.json").
		Query(api.Parameter{"screen_name": userID}).
		GET(ctx, api.OAuth1(e.ConsumerKey, e.AccessToken, e.SigningKey))

	if err != nil {
		return User{}, err
	}

	user := User{}
	err = mapstructure.Decode(resp, &user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (e *Endpoint) CheckFollowing(ctx context.Context, followingID string) (bool, error) {
	resp, err := api.New(ApiURL, "/1.1/friendships/show.json").
		Query(api.Parameter{
			"source_screen_name": e.UserID,
			"target_screen_name": followingID,
		}).
		GET(ctx, api.OAuth1(e.ConsumerKey, e.AccessToken, e.SigningKey))
	if err != nil {
		return false, err
	}

	if IsRateLimit(resp) {
		return false, ErrRateLimit
	}

	return resp.GetBool("relationship.source.following")
}
