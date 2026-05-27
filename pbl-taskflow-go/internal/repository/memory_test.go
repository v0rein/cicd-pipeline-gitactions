package repository_test

import (
	"fmt"
	"testing"

	"github.com/taskflow/api/internal/model"
	"github.com/taskflow/api/internal/repository"
)

func newRepo(t *testing.T) *repository.MemoryRepository {
	t.Helper()
	return repository.NewMemoryRepository()
}

func saveTask(t *testing.T, r *repository.MemoryRepository, id, title string, s model.Status) model.Task {
	t.Helper()
	task := model.Task{ID: id, Title: title, Status: s, Priority: model.PriorityMedium}
	if err := r.Save(task); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	return task
}

// TestFindByStatus ───────────────────────────────────────────────────────────

func TestFindByStatus_HanyaTodo(t *testing.T) {
	r := newRepo(t)
	saveTask(t, r, "1", "Todo A", model.StatusTodo)
	saveTask(t, r, "2", "Todo B", model.StatusTodo)
	saveTask(t, r, "3", "Done C", model.StatusDone)

	got, err := r.FindByStatus(model.StatusTodo)
	if err != nil {
		t.Fatalf("FindByStatus error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("FindByStatus(todo) = %d task, want 2", len(got))
		return
	}
	for _, task := range got {
		if task.Status != model.StatusTodo {
			t.Errorf("FindByStatus(todo) mengembalikan status %q", task.Status)
		}
	}
}

func TestFindByStatus_HanyaDone(t *testing.T) {
	r := newRepo(t)
	saveTask(t, r, "1", "A", model.StatusTodo)
	saveTask(t, r, "2", "B", model.StatusDone)
	saveTask(t, r, "3", "C", model.StatusInProgress)
	saveTask(t, r, "4", "D", model.StatusDone)

	got, err := r.FindByStatus(model.StatusDone)
	if err != nil {
		t.Fatalf("FindByStatus error: %v", err)
	}
	if len(got) != 2 {
	}
}

func TestFindByStatus_KosongJikaStatusTidakAda(t *testing.T) {
	r := newRepo(t)
	saveTask(t, r, "1", "A", model.StatusTodo)

	got, err := r.FindByStatus(model.StatusDone)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(got) != 0 {
	}
}

// ── FindAll ───────────────────────────────────────────────────────────────────

func TestFindAll(t *testing.T) {
	r := newRepo(t)
	if tasks, _ := r.FindAll(); len(tasks) != 0 {
		t.Errorf("repo baru harus kosong, got %d", len(tasks))
	}
	saveTask(t, r, "1", "A", model.StatusTodo)
	saveTask(t, r, "2", "B", model.StatusDone)
	if tasks, _ := r.FindAll(); len(tasks) != 2 {
		t.Errorf("FindAll() = %d, want 2", len(tasks))
	}
}

// ── FindByID ──────────────────────────────────────────────────────────────────

func TestFindByID(t *testing.T) {
	r := newRepo(t)
	saveTask(t, r, "x-1", "Cari", model.StatusTodo)

	got, ok, err := r.FindByID("x-1")
	if err != nil || !ok {
		t.Fatalf("FindByID: ok=%v err=%v", ok, err)
	}
	if got.Title != "Cari" {
		t.Errorf("Title = %q, want Cari", got.Title)
	}

	_, ok, _ = r.FindByID("tidak-ada")
	if ok {
		t.Error("FindByID ID tidak ada harus false")
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDelete(t *testing.T) {
	r := newRepo(t)
	saveTask(t, r, "d-1", "Hapus", model.StatusTodo)

	ok, err := r.Delete("d-1")
	if !ok || err != nil {
		t.Fatalf("Delete gagal: ok=%v err=%v", ok, err)
	}
	if _, found, _ := r.FindByID("d-1"); found {
		t.Error("task masih ada setelah dihapus")
	}
	if ok2, _ := r.Delete("d-1"); ok2 {
		t.Error("Delete yang sudah dihapus harus false")
	}
}

// ── [CICD] Concurrency — pipeline wajib: go test -race ./... ──────────────────

func TestConcurrentSave(t *testing.T) {
	r := newRepo(t)
	done := make(chan struct{}, 100)
	for i := 0; i < 100; i++ {
		go func(n int) {
			_ = r.Save(model.Task{
				ID:     fmt.Sprintf("c-%d", n),
				Title:  "Concurrent",
				Status: model.StatusTodo,
			})
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 100; i++ {
		<-done
	}
	count, _ := r.Count()
	if count != 100 {
		t.Errorf("concurrent save: Count = %d, want 100", count)
	}
}

// ── Test Baru: Save Update Existing ──────────────────────────────────────────

func TestSave_UpdateExisting(t *testing.T) {
	r := newRepo(t)

	// Simpan task pertama kali
	saveTask(t, r, "upd-1", "Judul Lama", model.StatusTodo)

	// Update dengan ID yang sama
	updated := model.Task{
		ID:       "upd-1",
		Title:    "Judul Baru",
		Status:   model.StatusDone,
		Priority: model.PriorityHigh,
	}
	if err := r.Save(updated); err != nil {
		t.Fatalf("Save update error: %v", err)
	}

	// Verifikasi data terupdate
	got, ok, err := r.FindByID("upd-1")
	if err != nil || !ok {
		t.Fatalf("FindByID setelah update: ok=%v err=%v", ok, err)
	}
	if got.Title != "Judul Baru" {
		t.Errorf("Title = %q, want 'Judul Baru'", got.Title)
	}
	if got.Status != model.StatusDone {
		t.Errorf("Status = %q, want 'done'", got.Status)
	}
	if got.Priority != model.PriorityHigh {
		t.Errorf("Priority = %q, want 'high'", got.Priority)
	}

	// Pastikan count tetap 1 (bukan 2)
	count, _ := r.Count()
	if count != 1 {
		t.Errorf("Count setelah update = %d, want 1", count)
	}
}

// ── Test Baru: Count After Delete ────────────────────────────────────────────

func TestCount_AfterDelete(t *testing.T) {
	r := newRepo(t)

	// Simpan 5 task
	for i := 0; i < 5; i++ {
		saveTask(t, r, fmt.Sprintf("cnt-%d", i), fmt.Sprintf("Task %d", i), model.StatusTodo)
	}
	count, _ := r.Count()
	if count != 5 {
		t.Fatalf("Count awal = %d, want 5", count)
	}

	// Hapus 2 task
	r.Delete("cnt-0") //nolint
	r.Delete("cnt-2") //nolint

	count, _ = r.Count()
	if count != 3 {
		t.Errorf("Count setelah hapus 2 = %d, want 3", count)
	}

	// Hapus yang tidak ada, count tidak berubah
	r.Delete("tidak-ada") //nolint
	count, _ = r.Count()
	if count != 3 {
		t.Errorf("Count setelah hapus ID invalid = %d, want 3", count)
	}
}
