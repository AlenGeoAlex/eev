package middleware

import (
	"backend-go/internal/httpx"
	"backend-go/internal/services"
	"context"
	"net/http"
)

// AutoRefreshMiddleware handles expired access tokens with cookie refresh
func AutoRefreshMiddleware(auth *services.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			accessCookie, err := r.Cookie(string(httpx.AccessTokenCookieKey))
			if err != nil {
				http.Error(w, "Unauthorized [NAT]", http.StatusUnauthorized)
				return
			}

			refreshCookie, err := r.Cookie(string(httpx.RefreshTokenCookieKey))
			if err != nil {
				http.Error(w, "Unauthorized [NRT]", http.StatusUnauthorized)
				return
			}

			claims, err := auth.ValidateAccessToken(accessCookie.Value, true)
			if err == nil {
				ctx := context.WithValue(r.Context(), httpx.UserClaimsKey, claims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			newAccess, newRefresh, _, err := auth.Refresh(r.Context(), accessCookie.Value, refreshCookie.Value)
			if err != nil {
				http.Error(w, "Unauthorized [FR]", http.StatusUnauthorized)
				return
			}

			// Set new cookies (HTTP-only, secure)
			http.SetCookie(w, &http.Cookie{
				Name:     string(httpx.AccessTokenCookieKey),
				Value:    newAccess,
				HttpOnly: true,
				Secure:   true,
				Path:     "/",
			})
			http.SetCookie(w, &http.Cookie{
				Name:     string(httpx.RefreshTokenCookieKey),
				Value:    newRefresh,
				HttpOnly: true,
				Secure:   true,
				Path:     "/",
			})

			newClaims, err := auth.ValidateAccessToken(newAccess, false)
			if err != nil {
				http.Error(w, "Unauthorized [FNT]", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), httpx.UserClaimsKey, newClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
