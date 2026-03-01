package httpx

import (
	"backend-go/config"
	_ "net/http"

	"github.com/go-chi/jwtauth/v5"
)

var tokenAuth *jwtauth.JWTAuth

func NewAuthMiddleware(cfg config.JwtConfig) *jwtauth.JWTAuth {
	tokenAuth = jwtauth.New("HS256", []byte(cfg.Secret), nil)
	return tokenAuth
}
