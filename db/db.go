package db

import (
	"JWTAuthentication/models"
)

var Users map[string]models.User

func init() {
	Users = make(map[string]models.User)
}
