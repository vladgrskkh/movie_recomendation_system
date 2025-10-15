package main

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken         = errors.New("Token is invalid")
	ErrNotEnoughTimeElapsed = errors.New("Not enough time elapsed before you can renew token")
)

var jwtKey = []byte("my secret key")

type claims struct {
	username string
	jwt.RegisteredClaims
}

// Maybe use a pointer to string
func createToken(username string) (string, error) {
	expireTime := time.Now().Add(24 * time.Hour)

	claims := claims{
		username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, err
}

func validateToken(token string) (*claims, error) {
	claims := &claims{}

	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !tkn.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// Maybe use a pointer to string
func refreshToken(token string) (string, error) {
	claims, err := validateToken(token)
	if err != nil {
		return "", err
	}

	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", ErrNotEnoughTimeElapsed
	}

	tokenString, err := createToken(claims.username)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
