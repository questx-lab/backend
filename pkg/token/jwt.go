package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/mitchellh/mapstructure"
)

type Engine interface {
	// Generate creates a token string containing the obj and expiration.
	Generate(expiration time.Duration, obj any) (string, error)

	// Verify if token is invalid or expired. Then parse the obj from token to obj parameter. The
	// obj paramter must be a pointer.
	Verify(token string, obj any) error
}

type standardClaims struct {
	jwt.RegisteredClaims
	Object any `json:"obj"`
}

type jwtEngine struct {
	secret string
}

func NewEngine(secret string) Engine {
	return &jwtEngine{secret: secret}
}

func (e *jwtEngine) Generate(expiration time.Duration, obj any) (string, error) {
	now := time.Now()
	claims := standardClaims{
		Object: obj,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(e.secret))
	return t, err
}

func (e *jwtEngine) Verify(token string, obj any) error {
	var claims standardClaims
	_, err := jwt.ParseWithClaims(
		token, &claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method %v", t.Header["alg"])
			}
			return []byte(e.secret), nil
		},
	)
	if err != nil {
		return err
	}

	err = mapstructure.Decode(claims.Object, obj)
	if err != nil {
		return err
	}

	return err
}
