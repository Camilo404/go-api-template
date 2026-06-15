// Package handlers contains the HTTP transport layer. Handlers are
// kept deliberately thin: parse the request, call a service, render
// the result. All cross-cutting concerns (logging, recovery, CORS)
// live in middleware.
package handlers

import (
	"net/http"
	"strings"

	"github.com/Camilo404/go-api-template/internal/httpx"
	"github.com/Camilo404/go-api-template/internal/models"
	"github.com/Camilo404/go-api-template/internal/service"
)

// TaskHandler is the example CRUD handler. Copy this file, rename
// the type, swap the service, and you have a new resource.
type TaskHandler struct {
	svc     *service.TaskService
	maxBody int64
}

// NewTaskHandler wires a TaskHandler with the given service and
// maximum request body size in bytes.
func NewTaskHandler(svc *service.TaskService, maxBody int64) *TaskHandler {
	return &TaskHandler{svc: svc, maxBody: maxBody}
}

// ServeHTTP dispatches /tasks and /tasks/{id} to the right method.
// The dispatch follows REST semantics: collection-only methods
// (POST) reject item paths, and item-only methods (PUT/PATCH/DELETE)
// reject the collection path. Either mismatch yields 405 instead
// of a generic 400.
func (h *TaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/tasks")
	path = strings.Trim(path, "/")
	isItem := path != ""

	switch r.Method {
	case http.MethodGet:
		if isItem {
			h.get(w, r, path)
		} else {
			h.list(w, r)
		}
	case http.MethodPost:
		if isItem {
			httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}
		h.create(w, r)
	case http.MethodPut, http.MethodPatch:
		if !isItem {
			httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}
		h.update(w, r, path)
	case http.MethodDelete:
		if !isItem {
			httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}
		h.delete(w, r, path)
	default:
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	}
}

// ListTasks godoc
// @Summary      List tasks
// @Description  Returns a paginated list of tasks. Defaults to page=1, per_page=20; per_page is capped at 100.
// @Tags         tasks
// @Produce      json
// @Param        page     query int false "Page number"   default(1) minimum(1)
// @Param        per_page query int false "Items per page" default(20) minimum(1) maximum(100)
// @Success      200 {object} models.PaginatedResponse
// @Failure      500 {object} models.APIError
// @Router       /api/v1/tasks [get]
func (h *TaskHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, perPage := models.ParsePagination(q.Get("page"), q.Get("per_page"), 20, 100)
	tasks, total, err := h.svc.List(r.Context(), page, perPage)
	if err != nil {
		httpx.WriteErrorFromErr(w, err)
		return
	}
	if tasks == nil {
		tasks = []models.Task{}
	}
	httpx.WriteJSON(w, http.StatusOK, models.PaginatedResponse{
		Data: tasks,
		Page: models.Page{Page: page, PerPage: perPage, Total: total},
	})
}

// GetTask godoc
// @Summary      Get a task
// @Description  Returns a single task by its identifier.
// @Tags         tasks
// @Produce      json
// @Param        id path string true "Task ID"
// @Success      200 {object} models.Task
// @Failure      400 {object} models.APIError
// @Failure      404 {object} models.APIError
// @Router       /api/v1/tasks/{id} [get]
func (h *TaskHandler) get(w http.ResponseWriter, r *http.Request, id string) {
	t, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpx.WriteErrorFromErr(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, t)
}

// CreateTask godoc
// @Summary      Create a task
// @Description  Creates a new task and returns the persisted entity.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        task body models.CreateTaskInput true "Task payload"
// @Success      201 {object} models.Task
// @Failure      400 {object} models.APIError
// @Router       /api/v1/tasks [post]
func (h *TaskHandler) create(w http.ResponseWriter, r *http.Request) {
	var in models.CreateTaskInput
	if err := httpx.DecodeJSON(w, r, h.maxBody, &in); err != nil {
		return // response already written
	}
	t, err := h.svc.Create(r.Context(), in)
	if err != nil {
		httpx.WriteErrorFromErr(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, t)
}

// UpdateTask godoc
// @Summary      Update a task
// @Description  Applies the provided fields. Both PUT and PATCH share the same payload; omitted fields are left unchanged.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id   path string                 true "Task ID"
// @Param        task body models.UpdateTaskInput true "Task payload"
// @Success      200 {object} models.Task
// @Failure      400 {object} models.APIError
// @Failure      404 {object} models.APIError
// @Router       /api/v1/tasks/{id} [put]
// @Router       /api/v1/tasks/{id} [patch]
func (h *TaskHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	var in models.UpdateTaskInput
	if err := httpx.DecodeJSON(w, r, h.maxBody, &in); err != nil {
		return
	}
	t, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		httpx.WriteErrorFromErr(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, t)
}

// DeleteTask godoc
// @Summary      Delete a task
// @Description  Permanently removes a task by its identifier.
// @Tags         tasks
// @Param        id path string true "Task ID"
// @Success      204
// @Failure      400 {object} models.APIError
// @Failure      404 {object} models.APIError
// @Router       /api/v1/tasks/{id} [delete]
func (h *TaskHandler) delete(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.Delete(r.Context(), id); err != nil {
		httpx.WriteErrorFromErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
