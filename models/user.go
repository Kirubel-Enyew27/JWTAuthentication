package models

import "github.com/dgrijalva/jwt-go"

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
}
