package handlers

import (
	"JWTAuthentication/customErrors"
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

var JWTKey []byte

func Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		panic(customErrors.UNABLE_TO_READ + "(error parsing request body)")
	}

	existingUser, ok := db.Users[user.Username]
	if !ok || bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password)) != nil {
		panic(customErrors.UNAUTHORIZED + "(invalid username or password)")
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
		panic(customErrors.UNABLE_TO_READ + "(" + err.Error() + ")")
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
		panic(customErrors.UNABLE_TO_SAVE + "(error parsing request body)")
	}

	if !isValidEmail(user.Email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if !isValidPhoneNumber(string(user.Phone)) {
		http.Error(w, "Invalid phone number format", http.StatusBadRequest)
		return
	}

	if !isValidAddress(user.Address) {
		http.Error(w, "Address cannot contain numbers", http.StatusBadRequest)
		return
	}

	if _, ok := db.Users[user.Username]; ok {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		panic(customErrors.UNABLE_TO_SAVE + "(unable to save user data)")
	}

	user.Password = string(hashedPassword)
	db.Users[user.Username] = user

	response := models.Response{
		MetaData: make(map[string]interface{}),
		Data:     user,
	}

	models.MetaDataHandler(w, response)
	w.Write([]byte("\n"))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully"))
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isValidPhoneNumber(phone string) bool {
	phoneRegex := regexp.MustCompile(`^(\+251|0)[79][0-9]{8}$`)
	return phoneRegex.MatchString(phone)
}

func isValidAddress(address string) bool {
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
	if err != nil {
		panic(customErrors.UNABLE_TO_READ + "(" + err.Error() + ")")
	}
	tokenString := tokenCookie.Value
	claims := &models.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JWTKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			panic(customErrors.UNAUTHORIZED + "(invalid token)")
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	claims.ExpiresAt = time.Now().Add(time.Hour).Unix()
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(JWTKey)
	if err != nil {
		panic(customErrors.UNABLE_TO_READ + "(" + err.Error() + ")")
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: time.Now().Add(time.Hour),
	})
}
