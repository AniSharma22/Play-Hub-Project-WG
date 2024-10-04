package errs

import (
	"encoding/json"
	"net/http"
)

type CustomError struct {
	error      `json:"-"`
	StatusCode int    `json:"-"`
	ErrorCode  string `json:"error_code"`
	Message    string `json:"message"`
}

func UnauthorizedError(errorCode string, message string) *CustomError {
	return &CustomError{
		StatusCode: http.StatusUnauthorized,
		ErrorCode:  errorCode,
		Message:    message,
	}
}

func ForbiddenError(errorCode string, message string) *CustomError {
	return &CustomError{
		StatusCode: http.StatusForbidden,
		ErrorCode:  errorCode,
		Message:    message,
	}
}

func InvalidRequestError(message string) *CustomError {
	return &CustomError{
		StatusCode: http.StatusBadRequest,
		ErrorCode:  "1100",
		Message:    message,
	}
}

func PermissionDeniedError(message string) *CustomError {
	return &CustomError{
		StatusCode: http.StatusUnauthorized,
		ErrorCode:  "2200",
		Message:    message,
	}
}

func ValidationError(message string) *CustomError {
	return &CustomError{
		StatusCode: http.StatusUnprocessableEntity,
		ErrorCode:  "3300",
		Message:    message,
	}
}

func DBError(message string) *CustomError {
	return &CustomError{
		StatusCode: http.StatusInternalServerError,
		ErrorCode:  "4400",
		Message:    message,
	}
}

func UnexpectedError(message string) *CustomError {
	return &CustomError{
		StatusCode: http.StatusInternalServerError,
		ErrorCode:  "9900",
		Message:    message,
	}
}

func (c *CustomError) ToJson2(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c.StatusCode)
	err := json.NewEncoder(w).Encode(c)
	if err != nil {
		// Handle encoding error (e.g., log it)
		return
	}
	return
}
