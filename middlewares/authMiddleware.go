package middlewares

import (
	"JWTAuthentication/handlers"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
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

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
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
