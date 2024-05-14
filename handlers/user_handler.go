package handlers

import (
	"JWTAuthentication/customErrors"
	"JWTAuthentication/db"
	"JWTAuthentication/models"
	"context"
	"io"
	"io/ioutil"
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
				Password: user.Password,
				Email:    user.Email,
				Phone:    user.Phone,
				Address:  user.Address,
			}
			users = append(users, u)
		}
		i++
	}

	if len(users) == 0 {
		error := customErrors.UNABLE_TO_FIND_RESOURCE + "(no users found)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
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
		error := customErrors.UNABLE_TO_READ + "(error parsing form data)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		error := customErrors.UNABLE_TO_READ + "(unable to read form data)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		error := customErrors.UNABLE_TO_READ + "(error reading file)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}

	contentType := http.DetectContentType(fileBytes)
	if !strings.HasPrefix(contentType, "image/") {
		error := customErrors.UNABLE_TO_READ + "(only image files are allowed)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}

	if err := os.MkdirAll("uploads", 0755); err != nil {
		error := customErrors.UNABLE_TO_SAVE + "(Error creating directory)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}

	f, err := os.OpenFile("uploads/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		error := customErrors.UNABLE_TO_SAVE + "(Error saving file)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}
	defer f.Close()

	if _, err := io.Copy(f, file); err != nil {
		error := customErrors.UNABLE_TO_SAVE + "(unable to save file)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
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
		error := customErrors.UNABLE_TO_FIND_RESOURCE + "(unable to find resource)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}
	defer file.Close()

	if _, err := io.Copy(w, file); err != nil {
		error := customErrors.UNABLE_TO_READ + "(unable to open file)"
		customErrors.CtxValue = context.WithValue(context.Background(), "errType", error)
		return
	}
}
