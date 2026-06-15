package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Camilo404/go-api-template/internal/models"
	"github.com/Camilo404/go-api-template/internal/service"
	"github.com/Camilo404/go-api-template/internal/store"
)

func newSvc() *service.TaskService {
	return service.NewTaskService(store.NewMemoryTaskStore())
}

func TestTaskService_Create_Validation(t *testing.T) {
	svc := newSvc()
	_, err := svc.Create(context.Background(), models.CreateTaskInput{Title: ""})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
	api, ok := err.(*models.APIError)
	if !ok || api.Code != models.ErrTitleRequired.Code {
		t.Errorf("expected title_required, got %v", err)
	}
}

func TestTaskService_Create_TrimsAndStores(t *testing.T) {
	svc := newSvc()
	got, err := svc.Create(context.Background(), models.CreateTaskInput{
		Title:       "  hello  ",
		Description: "  world  ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "hello" || got.Description != "world" {
		t.Errorf("expected trimmed values, got %+v", got)
	}
}

func TestTaskService_Update_InvalidID(t *testing.T) {
	svc := newSvc()
	_, err := svc.Update(context.Background(), "not-a-number", models.UpdateTaskInput{})
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected invalid_id, got %v", err)
	}
}

func TestTaskService_Update_EmptyTitle(t *testing.T) {
	svc := newSvc()
	created, _ := svc.Create(context.Background(), models.CreateTaskInput{Title: "ok"})

	empty := ""
	_, err := svc.Update(context.Background(), idAsString(created.ID), models.UpdateTaskInput{Title: &empty})
	if err == nil {
		t.Fatal("expected error for empty title on update")
	}
	var api *models.APIError
	if !errors.As(err, &api) || api.Code != models.ErrTitleRequired.Code {
		t.Errorf("expected title_required, got %v", err)
	}
}

func TestTaskService_FullFlow(t *testing.T) {
	svc := newSvc()
	ctx := context.Background()
	t1, err := svc.Create(ctx, models.CreateTaskInput{Title: "a"})
	if err != nil || t1.ID == 0 {
		t.Fatalf("create: %v %+v", err, t1)
	}
	if _, _, err := svc.List(ctx, 1, 10); err != nil {
		t.Fatalf("list: %v", err)
	}
	if err := svc.Delete(ctx, idAsString(t1.ID)); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

// idAsString is a tiny helper to keep the test file self-contained.
func idAsString(id int) string {
	const digits = "0123456789"
	if id == 0 {
		return "0"
	}
	out := make([]byte, 0, 8)
	for id > 0 {
		out = append([]byte{digits[id%10]}, out...)
		id /= 10
	}
	return string(out)
}
