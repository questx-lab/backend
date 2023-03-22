package testutil

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/jwt"
	"github.com/stretchr/testify/require"
)

func GetUserContext(t *testing.T) *api.Context {
	return GetUserContextWithAccessToken(t, model.AccessToken{
		ID:      "user1",
		Name:    "google",
		Address: "address",
	})
}

func GetUserContextWithAccessToken(t *testing.T, accessToken model.AccessToken) *api.Context {
	expiration := time.Minute
	engine := jwt.NewEngine[model.AccessToken]("secret", expiration)
	token, err := engine.Generate("", accessToken)
	require.Nil(t, err)

	req := &http.Request{Header: http.Header{}}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.AddCookie(&http.Cookie{Name: "access_name", Value: ""})

	return &api.Context{
		Cfg: config.Configs{
			Auth: config.AuthConfigs{
				AccessTokenName: "access_name",
				TokenSecret:     "secret",
				TokenExpiration: expiration,
			},
		},
		Request: req,
	}
}
