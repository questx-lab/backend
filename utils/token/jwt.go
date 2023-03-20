package token

import (
	"fmt"
	"time"

	"github.com/questx-lab/backend/config"

	"github.com/golang-jwt/jwt/v5"
)

func Generate(id string, configs *config.Configs) (string, error) {
	expirationTime := time.Now().Add(time.Duration(configs.JwtExpiredAt) * time.Second)
	claims := &jwt.RegisteredClaims{
		ID:        id,
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := jwtToken.SignedString([]byte(configs.JwtSecretKey))
	if err != nil {
		return "", fmt.Errorf("can not signed jwt: %w", err)
	}

	return token, nil

}

func Verify(token string, configs *config.Configs) (string, error) {
	keyFunc := func(jwtToken *jwt.Token) (interface{}, error) {

		if _, ok := jwtToken.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", fmt.Errorf("token is not valid")
		}
		return []byte(configs.JwtSecretKey), nil
	}
	claims := &jwt.RegisteredClaims{}
	if _, err := jwt.ParseWithClaims(token, claims, keyFunc); err != nil {
		return "", err
	}
	return claims.ID, nil
}
