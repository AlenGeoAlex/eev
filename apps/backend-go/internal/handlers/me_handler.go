package handlers

import (
	"backend-go/internal/httpx"
	"backend-go/internal/services"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type MeHandler struct {
	authService *services.AuthService
	userService *services.UserService
}

type MeResponse struct {
	ID     string  `json:"id"`
	Email  string  `json:"email"`
	Avatar *string `json:"avatar,omitempty"`
}

func NewMeHandler(authService *services.AuthService, userService *services.UserService) *MeHandler {
	return &MeHandler{
		authService: authService,
		userService: userService,
	}
}

func (receiver *MeHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(string(httpx.AccessTokenCookieKey))
	if err != nil {
		receiver.respondJSON(w, http.StatusUnauthorized, "Failed to get cookie")
		return
	}

	if cookie == nil {
		receiver.respondJSON(w, http.StatusUnauthorized, "No cookie found")
		return
	}

	token, err := receiver.authService.ValidateAccessToken(cookie.Value, true)
	if err != nil {
		receiver.respondJSON(w, http.StatusUnauthorized, "Failed to validate token")
		return
	}

	if token == nil {
		receiver.respondJSON(w, http.StatusUnauthorized, "No token found after validation")
		return
	}

	subject, err := token.GetSubject()
	if err != nil {
		receiver.respondJSON(w, http.StatusUnauthorized, "Failed to get subject from token")
		return
	}

	if subject == "" {
		receiver.respondJSON(w, http.StatusUnauthorized, "No subject found in token")
		return
	}

	user, err := receiver.userService.GetUserByID(r.Context(), uuid.MustParse(subject))
	if err != nil {
		receiver.respondJSON(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	if user == nil {
		receiver.respondJSON(w, http.StatusUnauthorized, "User not found")
		return
	}

	var avatar *string
	if user.AvatarUrl.Valid {
		avatar = &user.AvatarUrl.String
	}

	receiver.respondJSON(w, http.StatusOK, &MeResponse{
		ID:     user.ID,
		Email:  user.Email,
		Avatar: avatar,
	})
}

func (receiver *MeHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
