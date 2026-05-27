package service_test

import (
	"testing"

	"github.com/taskflow/api/internal/model"
	"github.com/taskflow/api/internal/repository"
	"github.com/taskflow/api/internal/service"
)

func newSvc() *service.TaskService {
	return service.NewTaskService(repository.NewMemoryRepository())
}

// TestCalculateCompletionRate ──────────────────────────────────────────────────

func TestCalculateCompletionRate(t *testing.T) {
	tests := []struct {
		name    string
		tasks   []model.Task
		want    float64
	}{
		{
			name:  "tidak ada task",
			tasks: []model.Task{},
			want:  0,
		},
		{
			name:  "semua done → 100%",
			tasks: []model.Task{{Status: model.StatusDone}, {Status: model.StatusDone}},
			want:  100.0,
		},
		{
			tasks: []model.Task{
				{Status: model.StatusDone},
				{Status: model.StatusTodo},
				{Status: model.StatusTodo},
			},
			want:  33.33,
		},
		{
			tasks: []model.Task{{Status: model.StatusDone}, {Status: model.StatusTodo}},
			want:  50.0,
		},
		{
			name: "tidak ada yang selesai → 0%",
			tasks: []model.Task{
				{Status: model.StatusTodo},
				{Status: model.StatusInProgress},
			},
			want: 0.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := service.CalculateCompletionRate(tc.tasks)
			// Toleransi 0.01 untuk floating point
			diff := got - tc.want
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.01 {
				t.Errorf("CalculateCompletionRate() = %.2f, want %.2f", got, tc.want)
			}
		})
	}
}

// ── Create ───────────────────────────────────────────────────────────────────

func TestCreate(t *testing.T) {
	svc := newSvc()

	t.Run("sukses dengan default priority", func(t *testing.T) {
		task, err := svc.Create(model.CreateTaskRequest{Title: "Belajar Go"})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if task.Title != "Belajar Go" {
			t.Errorf("Title = %q, want %q", task.Title, "Belajar Go")
		}
		if task.Status != model.StatusTodo {
			t.Errorf("Status = %q, want todo", task.Status)
		}
		if task.Priority != model.PriorityMedium {
			t.Errorf("Priority = %q, want medium (default)", task.Priority)
		}
		if task.ID == "" {
			t.Error("ID tidak boleh kosong")
		}
	})

	t.Run("title kosong ditolak", func(t *testing.T) {
		_, err := svc.Create(model.CreateTaskRequest{Title: ""})
		if err == nil {
			t.Error("Create() harus error jika title kosong")
		}
	})

	t.Run("title spasi saja ditolak", func(t *testing.T) {
		_, err := svc.Create(model.CreateTaskRequest{Title: "   "})
		if err == nil {
			t.Error("Create() harus error jika title hanya spasi")
		}
	})

	t.Run("priority invalid ditolak", func(t *testing.T) {
		_, err := svc.Create(model.CreateTaskRequest{Title: "T", Priority: "extreme"})
		if err == nil {
			t.Error("Create() harus error untuk priority tidak valid")
		}
	})

	t.Run("priority high sukses", func(t *testing.T) {
		task, err := svc.Create(model.CreateTaskRequest{Title: "Urgent", Priority: model.PriorityHigh})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if task.Priority != model.PriorityHigh {
			t.Errorf("Priority = %q, want high", task.Priority)
		}
	})

	t.Run("setiap task ID unik", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 50; i++ {
			task, _ := svc.Create(model.CreateTaskRequest{Title: "Task"})
			if ids[task.ID] {
				t.Errorf("ID duplikat ditemukan: %s", task.ID)
			}
			ids[task.ID] = true
		}
	})
}

// ── Update ───────────────────────────────────────────────────────────────────

func TestUpdate(t *testing.T) {
	svc := newSvc()

	t.Run("update status ke done mengisi completed_at", func(t *testing.T) {
		task, _ := svc.Create(model.CreateTaskRequest{Title: "Selesaikan"})
		statusDone := model.StatusDone
		updated, err := svc.Update(task.ID, model.UpdateTaskRequest{Status: &statusDone})
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if updated.CompletedAt == nil {
			t.Error("CompletedAt harus terisi setelah status = done")
		}
	})

	t.Run("update task tidak ada → error", func(t *testing.T) {
		statusDone := model.StatusDone
		_, err := svc.Update("id-tidak-ada", model.UpdateTaskRequest{Status: &statusDone})
		if err == nil {
			t.Error("Update() harus error untuk ID tidak ada")
		}
	})

	t.Run("update status invalid → error", func(t *testing.T) {
		task, _ := svc.Create(model.CreateTaskRequest{Title: "T"})
		s := model.Status("invalid")
		_, err := svc.Update(task.ID, model.UpdateTaskRequest{Status: &s})
		if err == nil {
			t.Error("Update() harus error untuk status tidak valid")
		}
	})
}

// ── [CICD] Full Task Lifecycle ────────────────────────────────────────────────
// [CICD] Simulasi integration test: create → get → update → delete.
// Jenis test ini dijalankan otomatis setelah deploy ke staging.

func TestTaskFullLifecycle(t *testing.T) {
	svc := newSvc()

	// 1. Create
	task, err := svc.Create(model.CreateTaskRequest{
		Title:    "Pipeline Lifecycle Test",
		Priority: model.PriorityHigh,
	})
	if err != nil {
		t.Fatalf("Create() gagal: %v", err)
	}

	// 2. Get
	got, err := svc.GetByID(task.ID)
	if err != nil || got.ID != task.ID {
		t.Fatalf("GetByID() gagal setelah create")
	}

	// 3. Update ke in_progress
	s := model.StatusInProgress
	got, err = svc.Update(task.ID, model.UpdateTaskRequest{Status: &s})
	if err != nil || got.Status != model.StatusInProgress {
		t.Fatalf("Update() ke in_progress gagal")
	}

	// 4. Update ke done
	done := model.StatusDone
	got, err = svc.Update(task.ID, model.UpdateTaskRequest{Status: &done})
	if err != nil || got.CompletedAt == nil {
		t.Fatalf("Update() ke done gagal atau CompletedAt nil")
	}

	// 5. Stats harus menunjukkan 1 done
	stats, err := svc.GetStats()
	if err != nil {
		t.Fatalf("GetStats() gagal: %v", err)
	}
	if stats.ByStatus["done"] != 1 {
		t.Errorf("Stats.ByStatus[done] = %d, want 1", stats.ByStatus["done"])
	}

	// 6. Delete
	_, err = svc.Delete(task.ID)
	if err != nil {
		t.Fatalf("Delete() gagal: %v", err)
	}

	// 7. Pastikan sudah terhapus
	if _, err = svc.GetByID(task.ID); err == nil {
		t.Error("GetByID() harus error setelah task dihapus")
	}
}

// ── [CICD] Rollback Simulation ───────────────────────────────────────────────

func TestRollbackStatusSimulation(t *testing.T) {
	svc := newSvc()
	task, _ := svc.Create(model.CreateTaskRequest{Title: "Rollback Test"})

	// Simulasi: deploy berhasil → update ke in_progress
	s := model.StatusInProgress
	svc.Update(task.ID, model.UpdateTaskRequest{Status: &s}) //nolint

	// Deployment bermasalah → rollback ke todo
	todo := model.StatusTodo
	rolled, err := svc.Update(task.ID, model.UpdateTaskRequest{Status: &todo})
	if err != nil {
		t.Fatalf("Rollback gagal: %v", err)
	}
	if rolled.Status != model.StatusTodo {
		t.Errorf("Setelah rollback, status = %q, want todo", rolled.Status)
	}
}

// ── Test Baru: Unicode Title ─────────────────────────────────────────────────

func TestCreate_WithUnicodeTitle(t *testing.T) {
	svc := newSvc()

	t.Run("title dengan emoji sukses", func(t *testing.T) {
		task, err := svc.Create(model.CreateTaskRequest{Title: "🚀 Deploy ke Production"})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if task.Title != "🚀 Deploy ke Production" {
			t.Errorf("Title = %q, want %q", task.Title, "🚀 Deploy ke Production")
		}
	})

	t.Run("title dengan karakter CJK sukses", func(t *testing.T) {
		task, err := svc.Create(model.CreateTaskRequest{Title: "タスク管理テスト"})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if task.Title != "タスク管理テスト" {
			t.Errorf("Title = %q, want %q", task.Title, "タスク管理テスト")
		}
	})
}

// ── Test Baru: Delete dan Verifikasi Stats ───────────────────────────────────

func TestDelete_AndVerifyStats(t *testing.T) {
	svc := newSvc()

	// Buat 3 task: 1 done, 2 todo
	t1, _ := svc.Create(model.CreateTaskRequest{Title: "Task 1"})
	t2, _ := svc.Create(model.CreateTaskRequest{Title: "Task 2"})
	svc.Create(model.CreateTaskRequest{Title: "Task 3"}) //nolint

	done := model.StatusDone
	svc.Update(t1.ID, model.UpdateTaskRequest{Status: &done}) //nolint

	// Stats awal: 3 total, 1 done
	stats, err := svc.GetStats()
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}
	if stats.Total != 3 {
		t.Errorf("Total = %d, want 3", stats.Total)
	}
	if stats.ByStatus["done"] != 1 {
		t.Errorf("ByStatus[done] = %d, want 1", stats.ByStatus["done"])
	}

	// Hapus task2 (todo)
	_, err = svc.Delete(t2.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Stats setelah delete: 2 total, 1 done, 1 todo
	stats, err = svc.GetStats()
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}
	if stats.Total != 2 {
		t.Errorf("Total setelah delete = %d, want 2", stats.Total)
	}
	if stats.ByStatus["done"] != 1 {
		t.Errorf("ByStatus[done] setelah delete = %d, want 1", stats.ByStatus["done"])
	}
	if stats.ByStatus["todo"] != 1 {
		t.Errorf("ByStatus[todo] setelah delete = %d, want 1", stats.ByStatus["todo"])
	}

	// CompletionRate: 1/2 = 50%
	diff := stats.CompletionRate - 50.0
	if diff < 0 {
		diff = -diff
	}
	if diff > 0.01 {
		t.Errorf("CompletionRate = %.2f, want 50.00", stats.CompletionRate)
	}
}

// ── Test Baru: GetAll ────────────────────────────────────────────────────────

func TestGetAll(t *testing.T) {
	svc := newSvc()
	
	// Create tasks
	svc.Create(model.CreateTaskRequest{Title: "Task A"})
	svc.Create(model.CreateTaskRequest{Title: "Task B"})
	
	tasks, err := svc.GetAll("")
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("GetAll() returned %d tasks, want 2", len(tasks))
	}
}
