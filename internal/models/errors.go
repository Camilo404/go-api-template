package models

import (
	"errors"
	"net/http"
)

// APIError is the wire format for all error responses. Using a typed
// error lets handlers map domain errors to HTTP status codes centrally.
type APIError struct {
	Code    string `json:"code" example:"not_found"`
	Message string `json:"message" example:"resource not found"`
}

// Error implements the error interface.
func (e *APIError) Error() string { return e.Message }

// NewAPIError constructs an APIError. Use the package-level sentinels
// below for common cases.
func NewAPIError(code, message string) *APIError {
	return &APIError{Code: code, Message: message}
}

// Common APIError sentinels. Add your own as you grow the service.
var (
	ErrNotFound           = NewAPIError("not_found", "resource not found")
	ErrTitleRequired      = NewAPIError("title_required", "title is required")
	ErrTitleTooLong       = NewAPIError("title_too_long", "title must be 200 characters or fewer")
	ErrDescriptionTooLong = NewAPIError("description_too_long", "description must be 2000 characters or fewer")
	ErrInvalidID          = NewAPIError("invalid_id", "invalid id")
	ErrInvalidBody        = NewAPIError("invalid_body", "invalid request body")
	ErrInternal           = NewAPIError("internal_error", "internal server error")
)

// HTTPStatus maps an APIError to its HTTP status code. Unknown errors
// are treated as 500.
func HTTPStatus(err error) int {
	var api *APIError
	if !errors.As(err, &api) {
		return http.StatusInternalServerError
	}
	switch api.Code {
	case ErrNotFound.Code:
		return http.StatusNotFound
	case ErrInvalidID.Code,
		ErrInvalidBody.Code,
		ErrTitleRequired.Code,
		ErrTitleTooLong.Code,
		ErrDescriptionTooLong.Code:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
