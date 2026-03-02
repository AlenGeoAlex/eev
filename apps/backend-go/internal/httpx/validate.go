package httpx

import (
	"backend-go/internal"
	"backend-go/internal/validation"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func ValidateBody[T any](next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var body T
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&body); err != nil {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err))
			return
		}

		if err := validation.Validate.Struct(body); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), BodyKey, body)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
