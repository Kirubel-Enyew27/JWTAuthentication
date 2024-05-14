package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func TimeOutMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestId := uuid.New().String()

		ctx := context.WithValue(r.Context(), "RequestID", requestId)
		ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
		defer cancel()

		done := make(chan struct{})
		defer close(done)

		var errType string

		go func() {
			defer func() {
				if err := recover(); err != nil {
					switch err.(type) {
					case string:
						errType = err.(string)
					default:
						errType = "Unknown Error"
					}
					ctx = context.WithValue(ctx, "ErrorType", errType)
				}
			}()

			next.ServeHTTP(w, r.WithContext(ctx))
			done <- struct{}{}
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			if errType != "" {
				http.Error(w, errType, http.StatusInternalServerError)
			} else {
				http.Error(w, "Request timed out", http.StatusGatewayTimeout)
			}
			return
		}
	}
}
