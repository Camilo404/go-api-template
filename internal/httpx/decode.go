package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Camilo404/go-api-template/internal/models"
)

// DecodeJSON reads, size-bounds, and decodes the request body into dst.
// Unknown JSON fields are rejected to surface client typos early. On
// error the response is already written; callers can simply return.
func DecodeJSON(w http.ResponseWriter, r *http.Request, maxBytes int64, dst any) error {
	if maxBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			WriteError(w, http.StatusRequestEntityTooLarge, "body_too_large", "request body too large")
			return err
		}
		WriteError(w, http.StatusBadRequest, models.ErrInvalidBody.Code, models.ErrInvalidBody.Message)
		return err
	}
	// Reject trailing garbage after the JSON document.
	if dec.More() {
		WriteError(w, http.StatusBadRequest, models.ErrInvalidBody.Code, "trailing data after JSON body")
		return errors.New("trailing data after JSON body")
	}
	return nil
}
