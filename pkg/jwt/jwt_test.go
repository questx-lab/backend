package jwt_test

import (
	"testing"
	"time"

	"github.com/questx-lab/backend/pkg/jwt"
	"github.com/stretchr/testify/require"
)

func TestJWT(t *testing.T) {
	engine := jwt.NewEngine[string]("secret", time.Minute)
	token, err := engine.Generate("", "abc")
	require.Nil(t, err)

	verifier := jwt.NewVerifier[string]("secret")
	msg, err := verifier.Verify(token)
	require.Nil(t, err)
	require.Equal(t, msg, "abc")
}

func TestJWTExpiration(t *testing.T) {
	engine := jwt.NewEngine[string]("secret", time.Nanosecond)
	token, err := engine.Generate("", "abc")
	require.Nil(t, err)

	verifier := jwt.NewVerifier[string]("secret")
	msg, err := verifier.Verify(token)
	require.NotNil(t, err)
	require.Equal(t, msg, "abc")
}

func TestJWTSameType(t *testing.T) {
	engine := jwt.NewEngine[string]("secret", time.Minute)
	token, err := engine.Generate("", "abc")
	require.Nil(t, err)

	verifier := jwt.NewVerifier[string]("secret")
	msg, err := verifier.Verify(token)
	require.Nil(t, err)
	require.Equal(t, msg, "abc")

	verifier = jwt.NewVerifier[string]("not secret")
	msg, err = verifier.Verify(token)
	require.NotNil(t, err)
	require.Equal(t, msg, "abc")
}

func TestJWTDiffType(t *testing.T) {
	engine := jwt.NewEngine[string]("secret", time.Minute)
	token, err := engine.Generate("", "abc")
	require.Nil(t, err)

	verifier := jwt.NewVerifier[int]("secret")
	_, err = verifier.Verify(token)
	require.NotNil(t, err)
}
