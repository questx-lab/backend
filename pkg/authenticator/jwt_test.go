package authenticator_test

import (
	"testing"
	"time"

	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/stretchr/testify/require"
)

func TestJWT(t *testing.T) {
	engine := authenticator.NewTokenEngine("secret")
	token, err := engine.Generate(time.Minute, "abc")
	require.Nil(t, err)

	var msg string
	err = engine.Verify(token, &msg)
	require.NoError(t, err)
	require.Equal(t, "abc", msg)
}

func TestJWTExpiration(t *testing.T) {
	engine := authenticator.NewTokenEngine("secret")
	token, err := engine.Generate(time.Nanosecond, "abc")
	require.Nil(t, err)

	var msg string
	err = engine.Verify(token, &msg)
	require.Error(t, err)
}
