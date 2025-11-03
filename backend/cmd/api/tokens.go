package main

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken         = errors.New("token is invalid")
	ErrNotEnoughTimeElapsed = errors.New("not enough time elapsed before you can renew token")
)

type Claims struct {
	UserID    int64 `json:"userID"`
	Activated bool  `json:"activated"`
	jwt.RegisteredClaims
}

func createToken(userID int64, activated bool, app *application) (string, error) {
	expireTime := time.Now().Add(30 * time.Minute)

	claims := Claims{
		UserID:    userID,
		Activated: activated,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(app.config.jwt.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// jwt.Validate
func validateToken(token string, app *application) (*Claims, error) {
	tkn, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (any, error) {
		return app.config.jwt.secretKey, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := tkn.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
