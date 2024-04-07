package handlers

import (
	"JWTAuthentication/db"
	"JWTAuthentication/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type PaginationParams struct {
	Page    int
	PerPage int
}

const (
	defaultMaxFileSize = 32 << 20
)

func parsePaginationParams(r *http.Request) PaginationParams {
	var params PaginationParams

	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("per_page")

	if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {
		params.Page = pageInt
	} else {
		params.Page = 1
	}

	if perPageInt, err := strconv.Atoi(perPageStr); err == nil && perPageInt > 0 {
		params.PerPage = perPageInt
	} else {
		params.PerPage = 3
	}

	return params
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	params := parsePaginationParams(r)
	offset := (params.Page - 1) * params.PerPage
	users := make([]models.User, 0, params.PerPage)

	i := 0
	for _, user := range db.Users {
		if i >= offset && len(users) < params.PerPage {
			users = append(users, user)
		}
		i++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, defaultMaxFileSize)
	if err := r.ParseMultipartForm(defaultMaxFileSize); err != nil {
		http.Error(w, "Error parsing form data", http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		fmt.Println("Error:", err)
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err := os.MkdirAll("uploads", 0755); err != nil {
		http.Error(w, "Error creating directory", http.StatusInternalServerError)
		return
	}

	f, err := os.OpenFile("uploads/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _, err := io.Copy(f, file); err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}

func GetImage(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/images/")
	file, err := os.Open("uploads/" + filename)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	if _, err := io.Copy(w, file); err != nil {
		http.Error(w, "Error serving file", http.StatusInternalServerError)
		return
	}
}
