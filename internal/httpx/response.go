// Package httpx contains small JSON HTTP helpers shared across handlers.
// Centralising response writing guarantees a consistent error shape and
// a single place to add features (e.g. compression, request tracing).
package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/Camilo404/go-api-template/internal/models"
)

// WriteJSON serialises v as JSON with the given status. The encoder is
// allowed to fail silently here because once headers are written there
// is nothing useful the handler can do.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError emits an APIError response.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, models.APIError{Code: code, Message: message})
}

// WriteErrorFromErr converts a domain error into the right HTTP
// response. Unknown errors become 500 with a generic message so we
// never leak internals to the client.
func WriteErrorFromErr(w http.ResponseWriter, err error) {
	if api, ok := err.(*models.APIError); ok && api != nil {
		WriteError(w, models.HTTPStatus(api), api.Code, api.Message)
		return
	}
	WriteError(w, http.StatusInternalServerError, models.ErrInternal.Code, models.ErrInternal.Message)
}
