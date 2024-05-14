package handlers

import (
	"JWTAuthentication/customErrors"
	"JWTAuthentication/db"
	"JWTAuthentication/models"
	"context"
	"encoding/json"
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
		error := customErrors.UNABLE_TO_READ + "(error parsing request body)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}

	existingUser, ok := db.Users[user.Username]
	if !ok || bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password)) != nil {
		error := customErrors.UNAUTHORIZED + "(invalid username or password)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
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
		error := customErrors.UNABLE_TO_READ + "(" + err.Error() + ")"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
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
		error := customErrors.UNABLE_TO_SAVE + "(error parsing request body)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}

	if user.Username == "" || user.Password == "" {
		error := customErrors.UNABLE_TO_SAVE + "(username and/or password cannot be empty)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}

	if !isValidEmail(user.Email) {
		error := customErrors.UNABLE_TO_SAVE + "(invalid email format)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}

	if !isValidPhoneNumber(string(user.Phone)) {
		error := customErrors.UNABLE_TO_SAVE + "(invalid phone number format)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}

	if !isValidAddress(user.Address) {
		error := customErrors.UNABLE_TO_SAVE + "(address cannot contain numbers)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}

	if _, ok := db.Users[user.Username]; ok {
		error := customErrors.UNABLE_TO_SAVE + "(username already exists)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		error := customErrors.UNABLE_TO_SAVE + "(unable to save user data)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
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
	phoneRegex := regexp.MustCompile(`^(\+251|251|0)?([79][0-9]{8})$`)
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
		error := customErrors.UNABLE_TO_READ + "(" + err.Error() + ")"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return

	}
	tokenString := tokenCookie.Value
	claims := &models.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JWTKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			error := customErrors.UNAUTHORIZED + "(invalid token)"
			customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
			return
		}
		error := customErrors.UNABLE_TO_READ + "(" + err.Error() + ")"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}

	claims.ExpiresAt = time.Now().Add(time.Hour).Unix()
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(JWTKey)
	if err != nil {
		error := customErrors.UNABLE_TO_READ + "(" + err.Error() + ")"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: time.Now().Add(time.Hour),
	})
}
