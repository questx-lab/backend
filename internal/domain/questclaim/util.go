package questclaim

import (
	"errors"
	"net/url"
	"strings"
)

type Tweet struct {
	TweetID        string
	UserScreenName string
}

type TwitterUser struct {
	UserScreenName string
}

func parseTweetURL(rawURL string) (Tweet, error) {
	path, err := getTwitterPath(rawURL)
	if err != nil {
		return Tweet{}, err
	}

	// The expected path is <user_id>/status/<tweet_id>
	parts := strings.Split(path, "/")
	if len(parts) != 3 || parts[1] != "status" {
		return Tweet{}, errors.New("invalid path")
	}

	return Tweet{TweetID: parts[2], UserScreenName: parts[0]}, nil
}

func parseTwitterUserURL(rawURL string) (TwitterUser, error) {
	path, err := getTwitterPath(rawURL)
	if err != nil {
		return TwitterUser{}, err
	}

	parts := strings.Split(path, "/")
	if len(parts) != 1 {
		return TwitterUser{}, errors.New("invalid path")
	}

	return TwitterUser{UserScreenName: parts[0]}, nil
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

func getDiscordInviteCode(rawURL string) (string, error) {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return "", err
	}

	if u.Scheme != "https" {
		return "", errors.New("invalid scheme")
	}

	if u.Host != "discord.gg" {
		return "", errors.New("invalid domain")
	}

	return strings.TrimLeft(u.Path, "/"), nil
}
