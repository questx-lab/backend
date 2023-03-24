package authenticator

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/questx-lab/backend/config"
)

type standardClaims[T any] struct {
	jwt.RegisteredClaims
	Object T `json:"obj,omitempty"`
}

type jwtTokenEngine[T any] struct {
	Expiration time.Duration

	secret string
}

func NewTokenEngine[T any](cfg config.TokenConfigs) TokenEngine[T] {
	return &jwtTokenEngine[T]{
		secret:     cfg.Secret,
		Expiration: cfg.Expiration,
	}
}

func (e *jwtTokenEngine[T]) Generate(sub string, obj T) (string, error) {
	now := time.Now()
	claims := standardClaims[T]{
		Object: obj,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(e.Expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   sub,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(e.secret))
	return t, err
}

func (e *jwtTokenEngine[T]) Verify(token string) (T, error) {
	var claims standardClaims[T]
	_, err := jwt.ParseWithClaims(
		token, &claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method %v", t.Header["alg"])
			}
			return []byte(e.secret), nil
		},
	)

	return claims.Object, err
}
