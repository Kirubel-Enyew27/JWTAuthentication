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
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", nil)

		defer func() {
			if err := customErrors.CtxValue.Value("errType"); err != nil {
				errMessage, _ := err.(string)
				errCode := strings.Split(errMessage, "(")[0]

				statusCode, ok := customErrors.ErrorStatusMap[errCode]
				if !ok {
					statusCode = http.StatusInternalServerError
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(statusCode)

				errorResponse := map[string]string{
					"error": errMessage,
				}
				json.NewEncoder(w).Encode(errorResponse)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
