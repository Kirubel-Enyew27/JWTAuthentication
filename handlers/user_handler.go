package handlers

import (
	"JWTAuthentication/db"
	"JWTAuthentication/models"
	"encoding/json"
	_ "fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func ListUsers(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("per_page")

	page := 1
	perPage := 3

	if pageStr != "" {
		pageInt, err := strconv.Atoi(pageStr)
		if err != nil || pageInt < 1 {
			http.Error(w, "Invalid page number", http.StatusBadRequest)
			return
		}
		page = pageInt
	}

	if perPageStr != "" {
		perPageInt, err := strconv.Atoi(perPageStr)
		if err != nil || perPageInt < 1 {
			http.Error(w, "Invalid per_page value", http.StatusBadRequest)
			return
		}
		perPage = perPageInt
	}

	offset := (page - 1) * perPage

	tokenCookie, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	tokenString := tokenCookie.Value
	claims := &models.Claims{}

	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JWTKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	if claims.ExpiresAt < time.Now().Unix() {
		http.Error(w, "Token expired", http.StatusUnauthorized)
		return
	}

	users := make([]models.User, 0, perPage)

	i := 0
	for _, user := range db.Users {
		if i >= offset && len(users) < perPage {
			users = append(users, user)
		}
		i++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func Upload(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("token")
	tokenString := tokenCookie.Value
	claims := &models.Claims{}

	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
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

	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	err = os.MkdirAll("uploads", 0755)
	if err != nil {
		http.Error(w, "Error creating directory", http.StatusInternalServerError)
		return
	}

	f, err := os.OpenFile("uploads/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}

func GetImage(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("token")
	tokenString := tokenCookie.Value
	claims := &models.Claims{}

	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
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

	filename := strings.TrimPrefix(r.URL.Path, "/images/")

	file, err := os.Open("uploads/" + filename)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Error serving file", http.StatusInternalServerError)
		return
	}
}
