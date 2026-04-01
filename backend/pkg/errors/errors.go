package apierrors

import (
	"errors"
	"net/http"
)

// Sentinel errors used across the application.
var (
	ErrNotFound        = errors.New("resource not found")
	ErrAlreadyExists   = errors.New("resource already exists")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrBadRequest      = errors.New("bad request")
	ErrInternalServer  = errors.New("internal server error")
	ErrInvalidInput    = errors.New("invalid input")
	ErrConflict        = errors.New("conflict")
	ErrStorageFailure  = errors.New("storage operation failed")
)

// APIError is the standard error envelope returned by all HTTP handlers.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *APIError) Error() string { return e.Message }

// New creates a new APIError.
func New(code int, message, details string) *APIError {
	return &APIError{Code: code, Message: message, Details: details}
}

// HTTPStatusFor maps a domain/sentinel error to an HTTP status code.
func HTTPStatusFor(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrAlreadyExists), errors.Is(err, ErrConflict):
		return http.StatusConflict
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrBadRequest), errors.Is(err, ErrInvalidInput):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
