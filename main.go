package main

import (
	"JWTAuthentication/handlers"
	"JWTAuthentication/middlewares"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		fmt.Println("JWT_KEY not found in environment variables")
		return
	}

	handlers.JWTKey = []byte(jwtKey)

	http.HandleFunc("/register", middlewares.TimeOutMiddleware(handlers.Register))
	http.HandleFunc("/login", middlewares.TimeOutMiddleware(handlers.Login))
	http.HandleFunc("/refresh", middlewares.TimeOutMiddleware(handlers.RefreshToken))
	http.HandleFunc("/users", middlewares.AuthMiddleware(middlewares.TimeOutMiddleware(handlers.ListUsers)))
	http.HandleFunc("/upload", middlewares.AuthMiddleware(middlewares.TimeOutMiddleware(handlers.Upload)))
	http.HandleFunc("/images/", middlewares.AuthMiddleware(middlewares.TimeOutMiddleware(handlers.GetImage)))

	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)

}
