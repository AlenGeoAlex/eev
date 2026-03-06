package handlers

import (
	"backend-go/internal"
	"backend-go/internal/httpx"
	"backend-go/internal/services"
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
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

type TargetUserEmailResponse struct {
	History []services.TargetUserEmails `json:"history"`
}

func NewMeHandler(authService *services.AuthService, userService *services.UserService) *MeHandler {
	return &MeHandler{
		authService: authService,
		userService: userService,
	}
}

// GetMe godoc
// @Summary Get current logged-in user info
// @Description Returns the currently authenticated user based on the access_token HttpOnly cookie.
// @Name Me
// @Tags user
// @Accept json
// @Produce json
// @Security CookieAuth
// @Success 200 {object} MeResponse "Successfully retrieved user"
// @Failure 401 {object} internal.ErrorResponse "Unauthorized - missing or invalid token"
// @Failure 500 {object} internal.ErrorResponse "Internal server error"
// @Router /me [get]
func (receiver *MeHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(string(httpx.AccessTokenCookieKey))
	if err != nil {
		receiver.respondError(w, http.StatusUnauthorized, "Failed to get cookie")
		return
	}

	if cookie == nil {
		receiver.respondError(w, http.StatusUnauthorized, "No cookie found")
		return
	}

	token, err := receiver.authService.ValidateAccessToken(cookie.Value, true)
	if err != nil {
		receiver.respondError(w, http.StatusUnauthorized, "Failed to validate token")
		return
	}

	if token == nil {
		receiver.respondError(w, http.StatusUnauthorized, "No token found after validation")
		return
	}

	subject, err := token.GetSubject()
	if err != nil {
		receiver.respondError(w, http.StatusUnauthorized, "Failed to get subject from token")
		return
	}

	if subject == "" {
		receiver.respondError(w, http.StatusUnauthorized, "No subject found in token")
		return
	}

	user, err := receiver.userService.GetUserByID(r.Context(), uuid.MustParse(subject))
	if err != nil {
		receiver.respondError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	if user == nil {
		receiver.respondError(w, http.StatusUnauthorized, "User not found")
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

// GetEmailHistory godoc
// @Summary Get the past email targets of the user
// @Description Returns the past email targets of the user
// @Name GetEmailHistory
// @Tags user
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param search query string false "Search email substring"
// @Success 200 {object} TargetUserEmailResponse "Successfully retrieved user history"
// @Failure 401 {object} internal.ErrorResponse "Unauthorized - missing or invalid token"
// @Failure 500 {object} internal.ErrorResponse "Internal server error"
// @Router /me/email-history [get]
func (receiver *MeHandler) GetEmailHistory(w http.ResponseWriter, r *http.Request) {
	context := r.Context()
	claims, ok := context.Value(httpx.UserClaimsKey).(*jwt.MapClaims)
	if !ok || claims == nil {
		receiver.respondError(w, http.StatusUnauthorized, "Failed to get claims")
		return
	}

	subject, err := claims.GetSubject()
	if err != nil || subject == "" {
		receiver.respondError(w, http.StatusUnauthorized, "Failed to get subject from claims")
		return
	}

	query := r.URL.Query()
	search := query.Get("q")
	subjectParsed, err := uuid.Parse(subject)
	if err != nil {
		receiver.respondError(w, http.StatusUnauthorized, "Failed to parse subject")
		return
	}

	user, err := receiver.userService.GetTargetEmailsOfUser(context, subjectParsed, &search)
	if err != nil {
		receiver.respondError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	receiver.respondJSON(w, http.StatusOK, user)
}

func (receiver *MeHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (receiver *MeHandler) respondError(w http.ResponseWriter, status int, message string) {
	receiver.respondJSON(w, status, internal.ErrorResponse{
		Message: message,
	})
}
