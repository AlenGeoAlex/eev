package handlers

import (
	"backend-go/config"
	"backend-go/internal"
	"backend-go/internal/httpx"
	"backend-go/internal/services"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type AuthHandler struct {
	appConfig   *config.AppConfig
	jwtConfig   *config.JwtConfig
	authService *services.AuthService
	userService *services.UserService
}

func NewAuthHandler(appConfig *config.AppConfig, authService *services.AuthService, userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		appConfig:   appConfig,
		jwtConfig:   &appConfig.Jwt,
		authService: authService,
		userService: userService,
	}
}

type GoogleCallbackRequest struct {
	Code  string `json:"code"  validate:"required"`
	State string `json:"state" validate:"required"`
}

type GoogleLoginResponse struct {
	URL string `json:"url"`
}

type GoogleCallbackResponse struct {
	ID     string  `json:"id"`
	Email  string  `json:"email"`
	Avatar *string `json:"avatar,omitempty"`
}

// GoogleLogin godoc
// @Name GoogleOAuthLogin
// @Summary Initiate Google OAuth login
// @Description Generates OAuth state, stores it in HttpOnly cookie, and returns Google authorization URL.
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} GoogleLoginResponse "Returns Google OAuth URL"
// @Failure 500 {object} internal.ErrorResponse "Failed to generate state"
// @Router /auth/google/login [get]
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := uuid.NewV7()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to generate state")
		return
	}

	strState := state.String()
	http.SetCookie(w, &http.Cookie{
		Name:     string(httpx.OAuthStateCookieKey),
		Value:    strState,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int((5 * time.Minute).Seconds()),
	})

	googleURL := h.authService.GetAuthURL(strState)
	h.respondJSON(w, http.StatusOK, GoogleLoginResponse{
		URL: googleURL,
	})
}

// GoogleCallback godoc
// @Name GoogleOAuthCallback
// @Summary Google OAuth callback
// @Description Validates Google OAuth code and state, creates or retrieves user, sets access_token and refresh_token HttpOnly cookies.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body GoogleCallbackRequest true "Google OAuth callback request"
// @Success 200 {object} GoogleCallbackResponse "User authenticated successfully"
// @Failure 400 {object} internal.ErrorResponse "Invalid request or state"
// @Failure 401 {object} internal.ErrorResponse "OAuth validation failed"
// @Failure 500 {object} internal.ErrorResponse "Internal server error"
// @Router /auth/google/callback [post]
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	body, ok := r.Context().Value(httpx.BodyKey).(GoogleCallbackRequest)
	if !ok {
		h.respondError(w, http.StatusBadRequest, "Failed to parse request body")
		return
	}

	cookie, err := r.Cookie(string(httpx.OAuthStateCookieKey))
	if err != nil || cookie.Value != body.State {
		h.respondError(w, http.StatusBadRequest, "Invalid state provided")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:   string(httpx.OAuthStateCookieKey),
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	gUser, err := h.authService.ValidateAndGetOAuthUser(r.Context(), body.Code)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, "OAuth validation failed")
		return
	}

	user, err := h.userService.GetUserByEmail(r.Context(), gUser.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			user, err = h.userService.CreateUser(r.Context(), gUser.Email, "google", &gUser.AvatarURL)
			if err != nil {
				h.respondError(w, http.StatusInternalServerError, "Failed to create user")
				return
			}
		} else {
			log.Println(err)
			h.respondError(w, http.StatusInternalServerError, "Failed to get user")
			return
		}
	}

	accessToken, refreshToken, expiry, err := h.authService.GenerateTokenPair(uuid.MustParse(user.ID), user.Email)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	secure := !h.appConfig.IsDevelopment()
	http.SetCookie(w, &http.Cookie{
		Name:     string(httpx.AccessTokenCookieKey),
		Value:    accessToken,
		HttpOnly: true,
		Secure:   secure,
		Path:     "/",
		Expires:  time.Now().Add(h.jwtConfig.AccessTTL),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     string(httpx.RefreshTokenCookieKey),
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   secure,
		Path:     "/",
		Expires:  expiry,
	})

	var avatar *string
	if user.AvatarUrl.Valid {
		avatar = &user.AvatarUrl.String
	}

	response := &GoogleCallbackResponse{
		ID:     user.ID,
		Email:  user.Email,
		Avatar: avatar,
	}

	h.respondJSON(w, http.StatusOK, response)
}

// Logout godoc
// @Name Logout
// @Summary Logout
// @Description Revokes tokens (if present) and clears cookies.
// @Tags auth
// @Accept json
// @Produce json
// @Success 204 "User logged out successfully"
// @Router /auth/logout [delete]
//func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
//	accessCookie, accessErr := r.Cookie(string(httpx.AccessTokenCookieKey))
//	refreshCookie, refreshErr := r.Cookie(string(httpx.RefreshTokenCookieKey))
//
//	jtiAccessToken := ""
//	jtiRefreshToken := ""
//	jwtUserId := ""
//
//	if accessErr == nil && accessCookie.Value != "" {
//		if claims, err := h.authService.ValidateAccessToken(accessCookie.Value, false); err == nil {
//			if userID, ok := (*claims)["sub"].(string); ok {
//				jwtUserId = userID
//			}
//
//			if jti, ok := (*claims)["jti"].(string); ok {
//				jtiAccessToken = jti
//			}
//		}
//	}
//
//	if refreshErr == nil && refreshCookie.Value != "" {
//		_ = h.authService.RevokeRefreshToken(r.Context(), refreshCookie.Value)
//	}
//
//	secure := !h.appConfig.IsDevelopment()
//	clearCookie := func(name string) {
//		http.SetCookie(w, &http.Cookie{
//			Name:     name,
//			Value:    "",
//			Path:     "/",
//			HttpOnly: true,
//			Secure:   secure,
//			MaxAge:   -1,
//		})
//	}
//
//	clearCookie(string(httpx.AccessTokenCookieKey))
//	clearCookie(string(httpx.RefreshTokenCookieKey))
//
//	w.WriteHeader(http.StatusNoContent)
//}

func (h *AuthHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *AuthHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, internal.ErrorResponse{
		Message: message,
	})
}
