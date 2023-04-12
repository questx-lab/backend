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

func (e *Endpoint) OnBehalf() string {
	return e.UserID
}

func (e *Endpoint) GetUser(ctx context.Context, userScreenName string) (User, error) {
	resp, err := api.New(ApiURL, "/1.1/users/show.json").
		Query(api.Parameter{"screen_name": userScreenName}).
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

func (e *Endpoint) GetTweet(ctx context.Context, tweetID string) (Tweet, error) {
	resp, err := api.New(ApiURL, "/1.1/statuses/show.json").
		Query(api.Parameter{"id": tweetID}).
		GET(ctx, api.OAuth1(e.ConsumerKey, e.AccessToken, e.SigningKey))

	if err != nil {
		return Tweet{}, err
	}

	id, err := resp.GetString("id_str")
	if err != nil {
		return Tweet{}, err
	}

	userScreenName, err := resp.GetString("user.screen_name")
	if err != nil {
		return Tweet{}, err
	}

	replyToTweetID, err := resp.GetString("in_reply_to_status_id_str")
	if err != nil {
		return Tweet{}, err
	}

	text, err := resp.GetString("text")
	if err != nil {
		return Tweet{}, err
	}

	return Tweet{
		ID:               id,
		AuthorScreenName: userScreenName,
		ReplyToTweetID:   replyToTweetID,
		Text:             text,
	}, nil
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

func (e *Endpoint) GetLikedTweet(ctx context.Context) ([]Tweet, error) {
	resp, err := api.New(ApiURL, "/1.1/favorites/list.json").
		Query(api.Parameter{"screen_name": e.UserID}).
		GET(ctx, api.OAuth1(e.ConsumerKey, e.AccessToken, e.SigningKey))

	if err != nil {
		return nil, err
	}

	array, err := resp.GetArray("array")
	if err != nil {
		return nil, err
	}

	var tweets []Tweet
	for _, x := range array {
		id, err := x.GetString("id_str")
		if err != nil {
			return nil, err
		}
		tweets = append(tweets, Tweet{ID: id})
	}

	return tweets, nil
}

func (e *Endpoint) GetRetweet(ctx context.Context, tweetID string) ([]Tweet, error) {
	resp, err := api.New(ApiURL, "/1.1/statuses/retweets/%s.json", tweetID).
		Query(api.Parameter{"count": "100"}).
		GET(ctx, api.OAuth1(e.ConsumerKey, e.AccessToken, e.SigningKey))

	if err != nil {
		return nil, err
	}

	array, err := resp.GetArray("array")
	if err != nil {
		return nil, err
	}

	var tweets []Tweet
	for _, tw := range array {
		id, err := tw.GetString("id_str")
		if err != nil {
			return nil, err
		}

		userScreenName, err := tw.GetString("user.screen_name")
		if err != nil {
			return nil, err
		}

		tweets = append(tweets, Tweet{ID: id, AuthorScreenName: userScreenName})
	}

	return tweets, nil
}
