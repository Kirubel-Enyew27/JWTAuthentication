package middlewares

import (
	"JWTAuthentication/customErrors"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func ErrorMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if r := recover(); r != nil {
				errMessage, _ := r.(string)
				errCode := strings.Split(errMessage, "(")[0]

				statusCode, ok := customErrors.ErrorStatusMap[errCode]
				if !ok {
					statusCode = http.StatusInternalServerError
				}

				ctx = context.WithValue(ctx, "error", errMessage)
				ctx = context.WithValue(ctx, "http_status", statusCode)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(statusCode)

				errorResponse := map[string]string{
					"error": errMessage,
				}
				json.NewEncoder(w).Encode(errorResponse)
			}
		}()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
