package authenticator

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/mitchellh/mapstructure"
)

type standardClaims struct {
	jwt.RegisteredClaims
	Object any `json:"obj"`
}

type jwtTokenEngine struct {
	secret string
}

func NewTokenEngine(secret string) TokenEngine {
	return &jwtTokenEngine{secret: secret}
}

func (e *jwtTokenEngine) Generate(expiration time.Duration, obj any) (string, error) {
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

func (e *jwtTokenEngine) Verify(token string, obj any) error {
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
