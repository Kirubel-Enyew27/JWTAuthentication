package models

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/dgrijalva/jwt-go"
)

type PhoneNumber string

func (p PhoneNumber) MarshalJSON() ([]byte, error) {
	var re = []*regexp.Regexp{
		regexp.MustCompile(`^(\+251|251|0)?([79][0-9]{8})$`),
		regexp.MustCompile(`^([79][0-9]{8})$`),
	}

	for _, r := range re {
		if match := r.FindStringSubmatch(string(p)); match != nil {
			formatted := "251" + match[2]
			return json.Marshal(formatted)
		}
	}

	return nil, errors.New("invalid phone number format")
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

type Response struct {
	MetaData map[string]interface{} `json:"meta_data"`
	Data     interface{}            `json:"data"`
}

func MetaDataHandler(w http.ResponseWriter, response Response) {

	if len(response.MetaData) == 0 {
		response = Response{
			MetaData: map[string]interface{}{
				"status":  "success",
				"message": "Request processed successfully",
			},
			Data: response.Data,
		}
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
