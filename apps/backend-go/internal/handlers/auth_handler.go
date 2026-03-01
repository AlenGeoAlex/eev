package handlers

import (
	"backend-go/config"
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

type GoogleCallbackResponse struct {
	ID     string  `json:"id"`
	Email  string  `json:"email"`
	Avatar *string `json:"avatar,omitempty"`
}

func (receiver *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := uuid.NewV7()
	if err != nil {
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
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

	googleURL := receiver.authService.GetAuthURL(strState)
	receiver.respondJSON(w, http.StatusOK, map[string]string{
		"url": googleURL,
	})
}

func (receiver *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	body, ok := r.Context().Value(httpx.BodyKey).(GoogleCallbackRequest)
	if !ok {
		http.Error(w, "missing request body", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie(string(httpx.OAuthStateCookieKey))
	if err != nil || cookie.Value != body.State {
		http.Error(w, "invalid state", http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:   string(httpx.OAuthStateCookieKey),
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	gUser, err := receiver.authService.ValidateAndGetOAuthUser(r.Context(), body.Code)
	if err != nil {
		http.Error(w, "oauth failed", http.StatusUnauthorized)
		return
	}

	user, err := receiver.userService.GetUserByEmail(r.Context(), gUser.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			user, err = receiver.userService.CreateUser(r.Context(), gUser.Email, "google", &gUser.AvatarURL)
			if err != nil {
				http.Error(w, "failed to create user", http.StatusInternalServerError)
				return
			}
		} else {
			log.Println(err)
			http.Error(w, "failed to get user", http.StatusInternalServerError)
			return
		}
	}

	accessToken, refreshToken, expiry, err := receiver.authService.GenerateTokenPair(uuid.MustParse(user.ID), user.Email)
	if err != nil {
		http.Error(w, "failed to generate tokens", http.StatusInternalServerError)
		return
	}

	secure := !receiver.appConfig.IsDevelopment()
	http.SetCookie(w, &http.Cookie{
		Name:     string(httpx.AccessTokenCookieKey),
		Value:    accessToken,
		HttpOnly: true,
		Secure:   secure,
		Path:     "/",
		Expires:  time.Now().Add(receiver.jwtConfig.AccessTTL),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     string(httpx.RefreshTokenCookieKey),
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
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

	receiver.respondJSON(w, http.StatusOK, response)
}

func (receiver *AuthHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
