package middleware

import (
	"context"
	"net/http"

	guid "github.com/satori/go.uuid"
)

const RequestIDContextKey = "requestID"

func ContextRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			uuid := guid.NewV4()
			reqID = uuid.String()
		}

		ctx = context.WithValue(ctx, RequestIDContextKey, reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
