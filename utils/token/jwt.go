package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtGenerator struct {
	configs *Configs
	name    string
}

func NewJWTGenerator(name string, configs *Configs) Generator {
	return &jwtGenerator{
		name:    name,
		configs: configs}
}

func (g *jwtGenerator) Generate(id string) (string, error) {
	expirationTime := time.Now().Add(time.Duration(g.configs.JwtExpiredAt) * time.Second)
	claims := &jwt.RegisteredClaims{
		ID:        id,
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := jwtToken.SignedString([]byte(g.configs.JwtSecretKey))
	if err != nil {
		return "", fmt.Errorf("can not signed jwt: %w", err)
	}

	return token, nil

}

func (g *jwtGenerator) Verify(token string) (string, error) {
	keyFunc := func(jwtToken *jwt.Token) (interface{}, error) {

		if _, ok := jwtToken.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", fmt.Errorf("token is not valid")
		}
		return []byte(g.configs.JwtSecretKey), nil
	}
	claims := &jwt.RegisteredClaims{}
	if _, err := jwt.ParseWithClaims(token, claims, keyFunc); err != nil {
		return "", err
	}
	return claims.ID, nil
}
