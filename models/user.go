package models

import (
	"encoding/json"

	"github.com/dgrijalva/jwt-go"
)

type PhoneNumber string

func (p PhoneNumber) MarshalJSON() ([]byte, error) {
	formatted := "2519" + string(p[3:])
	return json.Marshal(formatted)
}

type User struct {
	ID       string      `json:"id,omitempty"`
	Username string      `json:"username,omitempty"`
	Password string      `json:"password,omitempty"`
	Email    string      `json:"email,omitempty"`
	Phone    PhoneNumber `json:"phone,omitempty"`
	Address  string      `json:"address,omitempty"`
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
}
