package api

import (
	"net/http"
)

type oauth2Opt struct {
	token string
}

func OAuth2(prefix, token string) *oauth2Opt {
	return &oauth2Opt{token: prefix + " " + token}
}

func (opt *oauth2Opt) Do(client defaultClient, req *http.Request) {
	req.Header.Add("Authorization", opt.token)
}
