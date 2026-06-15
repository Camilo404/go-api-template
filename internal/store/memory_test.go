package store_test

import (
	"context"
	"testing"

	"github.com/Camilo404/go-api-template/internal/models"
	"github.com/Camilo404/go-api-template/internal/store"
)

func newStore() *store.MemoryTaskStore { return store.NewMemoryTaskStore() }

func TestMemoryTaskStore_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	s := newStore()

	created, err := s.Create(ctx, models.CreateTaskInput{Title: "buy milk"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID != 1 || created.Title != "buy milk" || created.Completed {
		t.Errorf("unexpected task: %+v", created)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Error("expected timestamps to be set")
	}

	got, err := s.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != created.Title {
		t.Errorf("got %q want %q", got.Title, created.Title)
	}
}

func TestMemoryTaskStore_ListPagination(t *testing.T) {
	ctx := context.Background()
	s := newStore()
	for i := 0; i < 25; i++ {
		if _, err := s.Create(ctx, models.CreateTaskInput{Title: "t"}); err != nil {
			t.Fatal(err)
		}
	}

	page1, total, err := s.List(ctx, 1, 10)
	if err != nil {
		t.Fatalf("list page 1: %v", err)
	}
	if total != 25 || len(page1) != 10 {
		t.Errorf("page1 len=%d total=%d", len(page1), total)
	}

	page3, total, err := s.List(ctx, 3, 10)
	if err != nil {
		t.Fatalf("list page 3: %v", err)
	}
	if total != 25 || len(page3) != 5 {
		t.Errorf("page3 len=%d total=%d", len(page3), total)
	}
}

func TestMemoryTaskStore_Update(t *testing.T) {
	ctx := context.Background()
	s := newStore()
	t1, _ := s.Create(ctx, models.CreateTaskInput{Title: "x"})

	completed := true
	newTitle := "y"
	upd, err := s.Update(ctx, t1.ID, models.UpdateTaskInput{
		Title:     &newTitle,
		Completed: &completed,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if upd.Title != "y" || !upd.Completed {
		t.Errorf("update: %+v", upd)
	}
}

func TestMemoryTaskStore_NotFound(t *testing.T) {
	ctx := context.Background()
	s := newStore()
	if _, err := s.GetByID(ctx, 999); err == nil {
		t.Error("expected error for missing id")
	}
	if err := s.Delete(ctx, 999); err == nil {
		t.Error("expected error deleting missing id")
	}
}

func TestMemoryTaskStore_Delete(t *testing.T) {
	ctx := context.Background()
	s := newStore()
	t1, _ := s.Create(ctx, models.CreateTaskInput{Title: "x"})

	if err := s.Delete(ctx, t1.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetByID(ctx, t1.ID); err == nil {
		t.Error("expected error after delete")
	}
}
