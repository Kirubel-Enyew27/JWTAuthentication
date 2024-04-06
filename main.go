package main

import (
	"JWTAuthentication/handlers"
	"context"
	"fmt"
	"net/http"
	"time"
)

func authenticateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func TimeOutMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		done := make(chan struct{})
		defer close(done)

		go func() {
			next.ServeHTTP(w, r.WithContext(ctx))
			done <- struct{}{}
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
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
