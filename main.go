package main

import (
	"fmt"
	"net/http"

	"JWTAuthentication/handlers"
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

func main() {
	http.HandleFunc("/login", handlers.Login)
	http.HandleFunc("/register", handlers.Register)
	http.HandleFunc("/refresh", handlers.RefreshToken)
	http.HandleFunc("/users", authenticateMiddleware(handlers.ListUsers))
	http.HandleFunc("/upload", authenticateMiddleware(handlers.Upload))
	http.HandleFunc("/images/", authenticateMiddleware(handlers.GetImage))

	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
