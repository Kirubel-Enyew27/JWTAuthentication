package customErrors

import (
	"context"
	"net/http"
)

const (
	UNABLE_TO_SAVE          = "UNABLE_TO_SAVE"
	UNABLE_TO_FIND_RESOURCE = "UNABLE_TO_FIND_RESOURCE"
	UNABLE_TO_READ          = "UNABLE_TO_READ"
	UNAUTHORIZED            = "UNAUTHORIZED"
)

var ErrorStatusMap = map[string]int{
	UNABLE_TO_SAVE:          http.StatusInternalServerError,
	UNABLE_TO_FIND_RESOURCE: http.StatusNotFound,
	UNABLE_TO_READ:          http.StatusInternalServerError,
	UNAUTHORIZED:            http.StatusUnauthorized,
}

var CtxValue context.Context
