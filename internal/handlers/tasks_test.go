package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Camilo404/go-api-template/internal/handlers"
	"github.com/Camilo404/go-api-template/internal/service"
	"github.com/Camilo404/go-api-template/internal/store"
)

func newHandler() http.Handler {
	svc := service.NewTaskService(store.NewMemoryTaskStore())
	return handlers.NewTaskHandler(svc, 1<<20)
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var rdr *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		rdr = bytes.NewReader(b)
	} else {
		rdr = bytes.NewReader(nil)
	}
	r := httptest.NewRequest(method, path, rdr)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestTaskHandler_CRUD(t *testing.T) {
	h := newHandler()

	// Create
	w := doJSON(t, h, http.MethodPost, "/tasks", map[string]string{"title": "alpha"})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: status=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"id":1`) {
		t.Errorf("create body missing id: %s", w.Body.String())
	}

	// List
	w = doJSON(t, h, http.MethodGet, "/tasks", nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "alpha") {
		t.Errorf("list: status=%d body=%s", w.Code, w.Body.String())
	}

	// Get
	w = doJSON(t, h, http.MethodGet, "/tasks/1", nil)
	if w.Code != http.StatusOK {
		t.Errorf("get: status=%d", w.Code)
	}

	// Update
	w = doJSON(t, h, http.MethodPut, "/tasks/1", map[string]any{"completed": true})
	if w.Code != http.StatusOK {
		t.Errorf("update: status=%d body=%s", w.Code, w.Body.String())
	}

	// Delete
	w = doJSON(t, h, http.MethodDelete, "/tasks/1", nil)
	if w.Code != http.StatusNoContent {
		t.Errorf("delete: status=%d", w.Code)
	}

	// Get missing
	w = doJSON(t, h, http.MethodGet, "/tasks/999", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("get missing: status=%d", w.Code)
	}
}

func TestTaskHandler_InvalidBody(t *testing.T) {
	h := newHandler()
	w := doJSON(t, h, http.MethodPost, "/tasks", map[string]string{"title": ""})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestTaskHandler_UnknownField(t *testing.T) {
	h := newHandler()
	w := doJSON(t, h, http.MethodPost, "/tasks", map[string]any{"title": "ok", "nope": 1})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown field, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestTaskHandler_MethodNotAllowed(t *testing.T) {
	h := newHandler()

	// PATCH on the collection is not allowed.
	w := doJSON(t, h, http.MethodPatch, "/tasks", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("PATCH /tasks: expected 405, got %d", w.Code)
	}

	// POST on a specific item is not allowed.
	w = doJSON(t, h, http.MethodPost, "/tasks/1", map[string]string{"title": "x"})
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /tasks/1: expected 405, got %d", w.Code)
	}

	// DELETE on the collection is not allowed.
	w = doJSON(t, h, http.MethodDelete, "/tasks", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("DELETE /tasks: expected 405, got %d", w.Code)
	}
}
