package handlers

import (
	"backend-go/config"
	"backend-go/internal"
	"backend-go/internal/httpx"
	"backend-go/internal/services"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ShareableHandler struct {
	shareableService *services.ShareableService
	appConfig        *config.AppConfig
}

type CreateShareableRequest struct {
	Name               *string                                   `json:"name" validate:"required"`
	Type               services.ShareableType                    `json:"type" validate:"required"`
	Data               *string                                   `json:"data"`
	Files              *[]ShareableFileRequest                   `json:"files"`
	AllowedEmails      *[]string                                 `json:"allowed_emails"`
	TimeParams         *CreateShareableRequestTimeParams         `json:"time_params" validate:"required"`
	NotificationParams *CreateShareableRequestNotificationParams `json:"notification_params"`
	Options            *CreateShareableRequestParams             `json:"options"`
}

type ShareableFileRequest struct {
	FileName    string `json:"file_name" validate:"required"`
	ContentType string `json:"content_type" validate:"required"`
	FileSize    int64  `json:"file_size" validate:"required"`
}

// ShareableFileUpload is what we return per file — the client uses upload_url to PUT directly to S3
type ShareableFileUpload struct {
	FileID    string `json:"file_id"`
	FileName  string `json:"file_name"`
	UploadURL string `json:"upload_url"`
}

type CreateShareableResponse struct {
	Code    string                `json:"code"`
	Uploads []ShareableFileUpload `json:"uploads,omitempty"` // only present for file type
}

type CreateShareableRequestParams struct {
	OnlyOnce bool `json:"only_once"`
	Encrypt  bool `json:"encrypt"`
}

type CreateShareableRequestNotificationParams struct {
	EmailNotificationOnOpen bool `json:"email_notification_on_open"`
	NotifyTargetEmails      bool `json:"notify_target_emails"`
}

type CreateShareableRequestTimeParams struct {
	ExpiryAt   *time.Time `json:"expiry_at"`
	ActiveFrom *time.Time `json:"active_from"`
}

type GetShareableResponse struct {
	Shareable services.ShareableCode          `json:"code"`
	Files     *[]services.ShareableSignedFile `json:"files"`
}

func NewShareableHandler(shareableService *services.ShareableService) *ShareableHandler {
	return &ShareableHandler{shareableService: shareableService}
}

// CreateShareable godoc
// @Summary Create a new shareable
// @Description Creates a new shareable resource (text, url, or file). Requires authentication via access_token cookie.
// @Name CreateShareable
// @Tags shareable
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body CreateShareableRequest true "Create shareable request"
// @Success 201 {object} CreateShareableResponse "Shareable created"
// @Success 202 {object} CreateShareableResponse "Shareable created - awaiting encrypted data"
// @Failure 400 {object} internal.ErrorResponse "Invalid request"
// @Failure 401 {object} internal.ErrorResponse "Unauthorized"
// @Failure 500 {object} internal.ErrorResponse "Internal server error"
// @Router /share [post]
func (h *ShareableHandler) CreateShareable(w http.ResponseWriter, r *http.Request) {
	body, ok := r.Context().Value(httpx.BodyKey).(CreateShareableRequest)
	if !ok {
		h.respondError(w, http.StatusBadRequest, "Failed to parse request body")
		return
	}

	if body.Type == services.ShareableTypeFile {
		if body.Data != nil && *body.Data != "" {
			h.respondError(w, http.StatusBadRequest, "Data must be empty for file type")
			return
		}
		if body.Files == nil || len(*body.Files) == 0 {
			h.respondError(w, http.StatusBadRequest, "At least one file is required for file type")
			return
		}

		contentSizeTotal := int64(0)
		for _, file := range *body.Files {
			contentSizeTotal += file.FileSize
		}

		if contentSizeTotal > int64(h.appConfig.MaxUploadSizeInMB)*int64(1024)*int64(1024) {
			h.respondError(w, http.StatusBadRequest, fmt.Sprintf("Total file size must be less than %d MB", h.appConfig.MaxUploadSizeInMB))
			return
		}
	} else {
		if body.Options.Encrypt && body.Data != nil {
			h.respondError(w, http.StatusBadRequest, "Data must be encrypted at client side for text and url shareables and patched to /api/share/{id}/encrypted")
			return
		}
	}

	if body.AllowedEmails == nil || len(*body.AllowedEmails) == 0 {
		body.AllowedEmails = nil
		if body.NotificationParams != nil {
			body.NotificationParams.NotifyTargetEmails = false
		}
	}

	activeFrom := time.Now().UTC()
	if body.TimeParams.ActiveFrom != nil {
		activeFrom = body.TimeParams.ActiveFrom.UTC()
	}

	expiryAt := time.Now().Add(time.Hour * 24).UTC()
	if body.TimeParams.ExpiryAt != nil {
		expiryAt = body.TimeParams.ExpiryAt.UTC()
	}

	if expiryAt.Compare(time.Now()) < 0 {
		h.respondError(w, http.StatusBadRequest, "Expiry must be in the future")
		return
	}

	if activeFrom.Compare(expiryAt) > 0 {
		h.respondError(w, http.StatusBadRequest, "Active from must be before expiry")
		return
	}

	options, err := h.buildOptions(body)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := h.userIDFromContext(r.Context())
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if body.Name == nil || *(body.Name) == "" {
		name := "Random Shareable"
		body.Name = &name
	}

	shareable, err := h.shareableService.CreateShareable(
		r.Context(),
		*body.Name, // If its null or empty, it will be set to "Random Shareable"
		userID,
		r.RemoteAddr,
		string(body.Type),
		body.Data,
		expiryAt,
		activeFrom,
		options,
	)
	if err != nil || shareable == nil {
		log.Printf("Failed to create shareable: %v", err)
		h.respondError(w, http.StatusInternalServerError, "failed to create shareable")
		return
	}

	if body.Type != services.ShareableTypeFile {
		resp := CreateShareableResponse{
			Code: shareable.ID,
		}
		if body.Options.Encrypt {
			h.respondJSON(w, http.StatusAccepted, resp)
		} else {
			h.respondJSON(w, http.StatusCreated, resp)
		}
		return
	}

	fileRequests := *(body.Files)
	totalFiles := len(fileRequests)
	fileUploads := make([]ShareableFileUpload, totalFiles)
	for i, request := range fileRequests {
		log.Printf("Processing file %d of %d", i+1, totalFiles)

		fileId, signedUrl, err := h.shareableService.CreateShareableFile(
			r.Context(),
			shareable.ID,
			shareable.UserID,
			request.FileName,
			request.FileSize,
			request.ContentType,
		)
		if err != nil {
			h.respondError(w, http.StatusInternalServerError, "Failed to create shareable file on "+request.FileName)
			// TODO DELETE SHAREABLE - SEND MESSAGE TO CHANNEL
			return
		}

		fileUploads[i] = ShareableFileUpload{
			FileName:  request.FileName,
			FileID:    fileId,
			UploadURL: signedUrl,
		}
		log.Printf("Prepared signed url for file %s to %s", request.FileName, signedUrl)
	}

	resp := CreateShareableResponse{
		Uploads: fileUploads,
		Code:    shareable.ID,
	}

	h.respondJSON(w, http.StatusCreated, resp)
}

func (h *ShareableHandler) buildOptions(body CreateShareableRequest) (map[services.ShareableOption]string, error) {
	opts := map[services.ShareableOption]string{}

	if body.Options != nil {
		if body.Options.OnlyOnce {
			opts[services.ShareableOptionOnlyOnce] = "true"
		}
		if body.Options.Encrypt {
			opts[services.ShareableOptionEncrypt] = "true"
			if body.Type != services.ShareableTypeFile {
				// Set flag to indicate that the encrypted data is awaiting to be client side encrypted
				// This will happen only for text and url shareables, files are already been taken care to be send from client
				opts[services.ShareableOptionAwaitingEncryptedData] = "true"
			}
		}
	}

	if body.AllowedEmails != nil {
		emailsJSON, err := json.Marshal(body.AllowedEmails)
		if err != nil {
			return nil, err
		}

		opts[services.ShareableOptionTargetEmails] = string(emailsJSON)
	}
	if body.NotificationParams != nil && body.NotificationParams.EmailNotificationOnOpen {
		opts[services.ShareableOptionEmailNotificationOnOpen] = "true"
	}

	if body.TimeParams.ActiveFrom != nil {
		opts[services.ShareableOptionActiveFrom] = body.TimeParams.ActiveFrom.Format(time.RFC3339)
	}

	return opts, nil
}

// GetShareable godoc
// @Summary Get shareable by code
// @Description Retrieves public shareable information using its unique code.
// @Tags shareable
// @Accept json
// @Produce json
// @Param code path string true "Shareable code"
// @Success 200 {object} services.ShareableCode "Shareable info"
// @Failure 400 {object} internal.ErrorResponse "Invalid code"
// @Failure 404 {object} internal.ErrorResponse "Shareable not found"
// @Failure 500 {object} internal.ErrorResponse "Internal server error"
// @Router /share/{code} [get]
func (h *ShareableHandler) GetShareable(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	if code == "" {
		http.Error(w, "invalid shareable code", http.StatusBadRequest)
		h.respondError(w, http.StatusBadRequest, "No shareable code provided in path")
		return
	}

	fromCode, err := h.shareableService.GetShareableInfoFromCode(code)
	if err != nil && fromCode == nil {
		h.respondError(w, http.StatusNotFound, "Shareable not found")
		return
	}
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if fromCode == nil {
		h.respondError(w, http.StatusNotFound, "Failed to get shareable info from code")
		return
	}

	response := GetShareableResponse{
		Shareable: *fromCode,
	}
	if fromCode.ShareableType == string(services.ShareableTypeFile) {
		files, err := h.shareableService.GetSignedFilesForShareable(fromCode.ID)
		if err != nil {
			h.respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response.Files = &files
	}

	h.respondJSON(w, http.StatusOK, response)
}

func (h *ShareableHandler) userIDFromContext(ctx context.Context) (string, error) {
	claims, ok := ctx.Value(httpx.UserClaimsKey).(*jwt.MapClaims)
	if !ok || claims == nil {
		return "", errors.New("no user claims found in context")
	}

	userIDRaw, ok := (*claims)["sub"]
	if !ok {
		return "", errors.New("sub missing in claims")
	}

	userID, ok := userIDRaw.(string)
	if !ok {
		return "", errors.New("sub is not a string")
	}

	return userID, nil
}

func (h *ShareableHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *ShareableHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, internal.ErrorResponse{
		Message: message,
	})
}
