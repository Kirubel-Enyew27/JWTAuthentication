package handlers

import (
	"JWTAuthentication/customErrors"
	"JWTAuthentication/db"
	"JWTAuthentication/models"
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
		params.PerPage = 5
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
			u := models.User{
				ID:       user.ID,
				Username: user.Username,
				Email:    user.Email,
				Phone:    user.Phone,
				Address:  user.Address,
			}
			users = append(users, u)
		}
		i++
	}

	if len(users) == 0 {
		panic(customErrors.UNABLE_TO_FIND_RESOURCE + "(No users found)")
	}

	for i := range users {
		if users[i].Phone == "" {
			users[i].Phone = models.PhoneNumber("")
		}
	}

	response := models.Response{
		MetaData: make(map[string]interface{}),
		Data:     users,
	}

	response.MetaData["Page"] = params.Page
	response.MetaData["PerPage"] = params.PerPage

	models.MetaDataHandler(w, response)
}

func Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, defaultMaxFileSize)
	if err := r.ParseMultipartForm(defaultMaxFileSize); err != nil {
		panic(customErrors.UNABLE_TO_READ + "(Error parsing form data)")
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		panic(customErrors.UNABLE_TO_READ + "(unable to read form data)")
	}
	defer file.Close()

	if err := os.MkdirAll("uploads", 0755); err != nil {
		panic(customErrors.UNABLE_TO_SAVE + "(Error creating directory)")
	}

	f, err := os.OpenFile("uploads/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(customErrors.UNABLE_TO_SAVE + "(Error saving file)")
	}
	defer f.Close()

	if _, err := io.Copy(f, file); err != nil {
		panic(customErrors.UNABLE_TO_SAVE + "(unable to save file)")
	}

	response := models.Response{
		MetaData: make(map[string]interface{}),
		Data:     "uploads/" + handler.Filename,
	}

	models.MetaDataHandler(w, response)
	w.Write([]byte("\n"))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}

func GetImage(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/images/")
	file, err := os.Open("uploads/" + filename)
	if err != nil {
		panic(customErrors.UNABLE_TO_FIND_RESOURCE + "(unable to find resource)")
	}
	defer file.Close()

	if _, err := io.Copy(w, file); err != nil {
		panic(customErrors.UNABLE_TO_READ + "(unable to open file)")
	}
}
