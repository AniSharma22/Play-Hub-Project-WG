package errs

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrEmailExists          = errors.New("email already exists")
	ErrGameNotFound         = errors.New("game not found")
	ErrDbError              = errors.New("db error")
	ErrAlreadyExists        = errors.New("already exists")
	ErrSlotPassed           = errors.New("slot passed")
	ErrUserAlreadyBooked    = errors.New("user already booked")
	ErrServiceError         = errors.New("service error")
	ErrSelfInviteError      = errors.New("can't invite self")
	ErrSlotFullyBookedError = errors.New("slot is already booked")
)

type AppError struct {
	error      `json:"-"`
	StatusCode int    `json:"error_code"`
	Message    string `json:"message"`
}

func NewNotFoundError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusNotFound,
		Message:    message,
	}
}

func NewBadRequestError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Message:    message,
	}
}

func NewInternalServerError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Message:    message,
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusConflict,
		Message:    message,
	}
}

func NewUnexpectedError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Message:    message,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusUnauthorized,
		Message:    message,
	}
}

func (e *AppError) ToJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.StatusCode)
	err := json.NewEncoder(w).Encode(e)
	if err != nil {
		// Handle encoding error (e.g., log it)
		return
	}
}
