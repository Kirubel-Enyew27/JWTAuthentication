package customErrors

import (
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

// func HandleHTTPError(w http.ResponseWriter, errCode string, errorMessage string) {
// 	statusCode, ok := errorStatusMap[errCode]
// 	if !ok {
// 		statusCode = http.StatusInternalServerError
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(statusCode)

// 	errorResponse := map[string]string{
// 		"error": errorMessage,
// 	}
// 	json.NewEncoder(w).Encode(errorResponse)
// }
