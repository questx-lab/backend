package discord

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/api"
	"github.com/stretchr/testify/require"
)

func Test_Endpoint_GiveRole_TooManyRequest(t *testing.T) {
	endpoint := New(context.Background(), config.DiscordConfigs{})

	resetAt := time.Now().Add(time.Second)
	endpoint.apiGenerator = &api.MockAPIGenerator{
		MockClient: api.MockAPIClient{
			PUTFunc: func(ctx context.Context, opts ...api.Opt) (*api.Response, error) {
				return &api.Response{
					Code:   http.StatusTooManyRequests,
					Header: http.Header{"X-Ratelimit-Reset": []string{strconv.FormatInt(resetAt.Unix(), 10)}},
				}, nil
			},
		},
	}

	// Call API with a response of TooManyRequest.
	err := endpoint.GiveRole(context.Background(), "guild-1", "user-1", "role-1")
	gotResetAt, ok := IsRateLimit(err)
	require.True(t, ok)
	require.Equal(t, resetAt.Unix(), gotResetAt.Unix())

	// Check the resource with identifier, ensure that it is limited.
	err = endpoint.checkLimitingResource(giveRoleResource, "guild-1")
	gotResetAt, ok = IsRateLimit(err)
	require.True(t, ok)
	require.Equal(t, resetAt.Unix(), gotResetAt.Unix())

	// Check another identifier, ensure that it is NOT limited.
	err = endpoint.checkLimitingResource(giveRoleResource, "guild-2")
	require.NoError(t, err)

	// Sleep until the limiting of resource expired. Check again.
	time.Sleep(time.Second)
	err = endpoint.checkLimitingResource(giveRoleResource, "guild-1")
	require.NoError(t, err)
}
