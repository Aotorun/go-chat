package api

import (
	"go-chat/internal/config"
	"go-chat/internal/domain"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func JWTMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(domain.JWTCustomClaims)
		},
		SigningKey:  []byte(cfg.JWTSecret),
		TokenLookup: "header:Authorization:Bearer ,query:token", // Искать токен в заголовке и в query-параметре "token" для WebSocket
		ErrorHandler: func(c echo.Context, err error) error {
			c.Logger().Errorf("JWT validation error: %v", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid or expired token",
			})
		},
	}
	return echojwt.WithConfig(config)
}
