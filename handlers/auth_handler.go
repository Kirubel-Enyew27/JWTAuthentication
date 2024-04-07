package handlers

import (
	"JWTAuthentication/db"
	"JWTAuthentication/models"
	_ "context"
	"encoding/json"
	_ "fmt"
	"net/http"
	"regexp"
	"time"
	"unicode"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var JWTKey = []byte("kcccck")

func Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	existingUser, ok := db.Users[user.Username]
	if !ok || bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password)) != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(time.Hour)
	claims := &models.Claims{
		UserID:   existingUser.ID,
		Username: existingUser.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JWTKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
	w.Write([]byte("User logged in successfully"))
}

func Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate email format
	if !isValidEmail(user.Email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Validate phone number format
	if !isValidPhoneNumber(user.Phone) {
		http.Error(w, "Invalid phone number format", http.StatusBadRequest)
		return
	}

	// Validate address field
	if !isValidAddress(user.Address) {
		http.Error(w, "Address cannot contain numbers except if it's empty", http.StatusBadRequest)
		return
	}

	if _, ok := db.Users[user.Username]; ok {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user.Password = string(hashedPassword)
	db.Users[user.Username] = user

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully"))
}

func isValidEmail(email string) bool {
	// Basic email format validation using regex
	// This regex may not cover all valid email formats
	// Adjust it according to your specific requirements
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isValidPhoneNumber(phone string) bool {
	// Basic phone number format validation using regex
	// This regex may not cover all valid phone number formats
	// Adjust it according to your specific requirements
	phoneRegex := regexp.MustCompile(`^(\+251|0)[79][0-9]{8}$`)
	return phoneRegex.MatchString(phone)
}

func isValidAddress(address string) bool {
	// Validate address field to ensure it does not contain any numbers
	// except if it's empty
	if address == "" {
		return true
	}

	for _, char := range address {
		if unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("token")
	tokenString := tokenCookie.Value
	claims := &models.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JWTKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims.ExpiresAt = time.Now().Add(time.Hour).Unix()
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(JWTKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: time.Now().Add(time.Hour),
	})
}
