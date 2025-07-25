package domain

import "github.com/golang-jwt/jwt/v5"

type JWTCustomClaims struct {
	UserID int64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
