package jwt_test

import (
	"testing"
	"time"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/jwt"
	"github.com/stretchr/testify/require"
)

func TestJWT(t *testing.T) {
	engine := jwt.NewEngine[string](config.TokenConfigs{Secret: "secret", Expiration: time.Minute})
	token, err := engine.Generate("", "abc")
	require.Nil(t, err)

	msg, err := engine.Verify(token)
	require.Nil(t, err)
	require.Equal(t, msg, "abc")
}

func TestJWTExpiration(t *testing.T) {
	engine := jwt.NewEngine[string](config.TokenConfigs{Secret: "secret", Expiration: time.Nanosecond})
	token, err := engine.Generate("", "abc")
	require.Nil(t, err)

	msg, err := engine.Verify(token)
	require.NotNil(t, err)
	require.Equal(t, msg, "abc")
}

func TestJWTSameType(t *testing.T) {
	engine := jwt.NewEngine[string](config.TokenConfigs{Secret: "secret", Expiration: time.Minute})
	token, err := engine.Generate("", "abc")
	require.Nil(t, err)

	msg, err := engine.Verify(token)
	require.Nil(t, err)
	require.Equal(t, msg, "abc")

	engine = jwt.NewEngine[string](config.TokenConfigs{Secret: "not secret", Expiration: time.Minute})
	msg, err = engine.Verify(token)
	require.NotNil(t, err)
	require.Equal(t, msg, "abc")
}

func TestJWTDiffType(t *testing.T) {
	engineStr := jwt.NewEngine[string](config.TokenConfigs{Secret: "secret", Expiration: time.Minute})
	token, err := engineStr.Generate("", "abc")
	require.Nil(t, err)

	engineInt := jwt.NewEngine[int](config.TokenConfigs{Secret: "secret", Expiration: time.Minute})
	_, err = engineInt.Verify(token)
	require.NotNil(t, err)
}
