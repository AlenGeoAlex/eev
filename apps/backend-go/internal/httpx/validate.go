package httpx

import (
	"backend-go/internal/validation"
	"context"
	"encoding/json"
	"log"
	"net/http"
)

func ValidateBody[T any](next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var body T
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&body); err != nil {
			log.Println(err)
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		if err := validation.Validate.Struct(body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), BodyKey, body)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
