package main

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken         = errors.New("token is invalid")
	ErrNotEnoughTimeElapsed = errors.New("not enough time elapsed before you can renew token")
)

// Placeholder will change it later
var jwtKey = []byte("my secret key")

type Claims struct {
	UserID int64  `json:"userID"`
	Scope  string `json:"scope"`
	jwt.RegisteredClaims
}

// Maybe use a pointer to string
func createTokenAuth(userID int64) (string, error) {
	expireTime := time.Now().Add(24 * time.Hour)

	claims := Claims{
		UserID: userID,
		Scope:  "authorization",
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

func createTokenActivation(userID int64) (string, error) {
	expireTime := time.Now().Add(24 * time.Hour)

	claims := Claims{
		UserID: userID,
		Scope:  "activation",
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

func validateToken(token string) (*Claims, error) {
	tkn, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !tkn.Valid {
		return nil, ErrInvalidToken
	}
	claims, ok := tkn.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	slog.Info(claims.Scope)
	slog.Info(strconv.Itoa(int(claims.UserID)))

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

	tokenString, err := createTokenAuth(claims.UserID)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
