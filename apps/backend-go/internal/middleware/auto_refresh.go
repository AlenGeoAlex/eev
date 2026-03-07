package middleware

import (
	"backend-go/internal"
	"backend-go/internal/httpx"
	"backend-go/internal/services"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// AutoRefreshMiddleware handles expired access tokens with cookie refresh
func AutoRefreshMiddleware(auth *services.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			accessCookie, err := r.Cookie(string(httpx.AccessTokenCookieKey))
			if err != nil {
				respondError(w, http.StatusUnauthorized, "Unauthorized [NAT]")
				return
			}

			refreshCookie, err := r.Cookie(string(httpx.RefreshTokenCookieKey))
			if err != nil {
				respondError(w, http.StatusUnauthorized, "Unauthorized [NRT]")
				return
			}

			claims, err := auth.ValidateAccessToken(accessCookie.Value, true)
			if err == nil {
				ctx := context.WithValue(r.Context(), httpx.UserClaimsKey, claims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if !errors.Is(err, services.ErrTokenExpired) {
				respondError(w, http.StatusUnauthorized, "Unauthorized [TAMPERING]")
				return
			}

			newAccess, newRefresh, _, err := auth.Refresh(r.Context(), accessCookie.Value, refreshCookie.Value)
			if err != nil {
				log.Println("Refresh failed", err)
				respondError(w, http.StatusUnauthorized, "Unauthorized [FR]")
				return
			}

			log.Println("Refresh succeeded")

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
				respondError(w, http.StatusUnauthorized, "Unauthorized [FNT]")
				return
			}

			ctx := context.WithValue(r.Context(), httpx.UserClaimsKey, newClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, internal.ErrorResponse{
		Message: message,
	})
}
