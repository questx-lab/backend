package questclaim

import (
	"errors"
	"net/url"
	"strings"
)

type TwitterUser struct {
	UserScreenID string
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

	return TwitterUser{UserScreenID: parts[0]}, nil
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
