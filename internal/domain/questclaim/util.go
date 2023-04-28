package questclaim

import (
	"errors"
	"net/url"
	"strings"
)

type tweet struct {
	TweetID        string
	UserScreenName string
}

type twitterUser struct {
	UserScreenName string
}

func parseTweetURL(rawURL string) (tweet, error) {
	path, err := getTwitterPath(rawURL)
	if err != nil {
		return tweet{}, err
	}

	// The expected path is <user_id>/status/<tweet_id>
	parts := strings.Split(path, "/")
	if len(parts) != 3 || parts[1] != "status" {
		return tweet{}, errors.New("invalid path")
	}

	return tweet{TweetID: parts[2], UserScreenName: parts[0]}, nil
}

func parseTwitterUserURL(rawURL string) (twitterUser, error) {
	path, err := getTwitterPath(rawURL)
	if err != nil {
		return twitterUser{}, err
	}

	parts := strings.Split(path, "/")
	if len(parts) != 1 {
		return twitterUser{}, errors.New("invalid path")
	}

	return twitterUser{UserScreenName: parts[0]}, nil
}

func getTwitterPath(rawURL string) (string, error) {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return "", err
	}

	if u.Scheme != "https" {
		return "", errors.New("invalid scheme")
	}

	if u.Host != "twitter.com" {
		return "", errors.New("invalid domain")
	}

	return strings.TrimLeft(u.Path, "/"), nil
}

func parseInviteTelegramURL(rawURL string) (string, error) {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return "", err
	}

	if u.Scheme != "https" {
		return "", errors.New("invalid scheme")
	}

	if u.Host != "t.me" {
		return "", errors.New("invalid domain")
	}

	return strings.TrimLeft(u.Path, "/"), nil
}
