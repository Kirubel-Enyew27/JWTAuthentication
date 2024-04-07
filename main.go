package main

import (
	"JWTAuthentication/handlers"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

const (
	UNABLE_TO_SAVE          = "UNABLE_TO_SAVE"
	UNABLE_TO_FIND_RESOURCE = "UNABLE_TO_FIND_RESOURCE"
	UNABLE_TO_READ          = "UNABLE_TO_READ"
	UNAUTHORIZED            = "UNAUTHORIZED"
)

func extractUserIDFromToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(handlers.JWTKey), nil
	})
	if err != nil {
		return "", fmt.Errorf("token parsing error: %v", err)
	}

	if !token.Valid {
		return "", errors.New("token is invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("unable to extract claims: claims are not of type jwt.MapClaims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", errors.New("unable to extract user ID: UserID claim not found or not a string")
	}
	return userID, nil
}

func authenticateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		token := tokenCookie.Value

		userId, err := extractUserIDFromToken(token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "UserID", userId)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func TimeOutMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestId := uuid.New().String()

		ctx := context.WithValue(r.Context(), "RequestID", requestId)
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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

func main() {

	http.HandleFunc("/register", TimeOutMiddleware(handlers.Register))
	http.HandleFunc("/login", TimeOutMiddleware(handlers.Login))
	http.HandleFunc("/refresh", TimeOutMiddleware(handlers.RefreshToken))
	http.HandleFunc("/users", authenticateMiddleware(TimeOutMiddleware(handlers.ListUsers)))
	http.HandleFunc("/upload", authenticateMiddleware(TimeOutMiddleware(handlers.Upload)))
	http.HandleFunc("/images/", authenticateMiddleware(TimeOutMiddleware(handlers.GetImage)))

	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)

}
