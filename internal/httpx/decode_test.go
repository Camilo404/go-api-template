package httpx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSON_Success(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":1}`))
	w := httptest.NewRecorder()
	var dst struct {
		A int `json:"a"`
	}
	if err := DecodeJSON(w, r, 1<<10, &dst); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if dst.A != 1 {
		t.Errorf("got %d", dst.A)
	}
}

func TestDecodeJSON_RejectsUnknownField(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":1,"b":2}`))
	w := httptest.NewRecorder()
	var dst struct {
		A int `json:"a"`
	}
	err := DecodeJSON(w, r, 1<<10, &dst)
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDecodeJSON_BodyTooLarge(t *testing.T) {
	big := strings.Repeat("x", 100)
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":"`+big+`"}`))
	w := httptest.NewRecorder()
	var dst struct {
		A string `json:"a"`
	}
	err := DecodeJSON(w, r, 16, &dst)
	if err == nil {
		t.Fatal("expected error for oversize body")
	}
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", w.Code)
	}
	if !errors.Is(err, err) { // sanity: error returned even though response is written
		t.Error("expected non-nil error")
	}
}
