//go:build integration

package management

import (
	"context"
	"errors"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/kar1hsu/go-seed/pkg/storage"
	"github.com/kar1hsu/go-seed/services/system/internal/application/jobs"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/internal/task"
	"github.com/kar1hsu/go-seed/services/system/resources"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func fileJobs(t *testing.T, f *fixture) (storage.Store, *jobs.Service) {
	t.Helper()
	objects, err := storage.New(storage.Config{Driver: "local", Root: t.TempDir()})
	require(t, err)
	f.a.objects, f.b.objects = objects, objects
	return objects, jobs.New(repository.NewFilesTasks(f.db), repository.NewRBAC(f.db), f.a.authorizer, objects)
}
func TestFilesPrivacyVerificationAndDelete(t *testing.T) {
	f := setup(t)
	objects, worker := fileJobs(t, f)
	_, other := f.user(t, "file_other")
	record, err := f.a.CreateFile(ctx, f.root, "readme.txt", "text/plain", 5)
	require(t, err)
	if !strings.HasPrefix(record.ObjectKey, record.CreatedAt.UTC().Format("20060102/")) || len(record.ObjectKey) != 41 {
		t.Fatalf("upload key does not use its creation date: %q", record.ObjectKey)
	}
	sameName, err := f.a.CreateFile(ctx, f.root, "readme.txt", "text/plain", 5)
	require(t, err)
	if sameName.ObjectKey == record.ObjectKey {
		t.Fatal("same original filename reused an object key")
	}
	// Existing flat keys keep working after the column is widened.
	sameName.ObjectKey = strings.Repeat("c", 32)
	require(t, f.db.Model(sameName).Update("object_key", sameName.ObjectKey).Error)
	require(t, objects.Put(ctx, sameName.ObjectKey, strings.NewReader("older"), 5))
	_, err = f.a.CompleteFileUpload(ctx, f.root, sameName.ID)
	require(t, err)
	_, err = f.a.DownloadFile(ctx, f.root, sameName.ID)
	require(t, err)
	require(t, f.a.DeleteFile(ctx, f.root, sameName.ID))
	var legacyDelete model.Task
	require(t, f.db.Where("file_id = ?", sameName.ID).First(&legacyDelete).Error)
	require(t, worker.Process(ctx, task.Payload{ID: legacyDelete.ID, Generation: legacyDelete.Generation}))
	if _, err = objects.Open(ctx, sameName.ObjectKey); !os.IsNotExist(err) {
		t.Fatal("legacy object was not deleted", err)
	}
	_, err = f.a.AuthorizeFileUpload(ctx, other, record.ID)
	code(t, err, codes.PermissionDenied)
	require(t, objects.Put(ctx, record.ObjectKey, strings.NewReader("hello"), 5))
	ready, err := f.a.CompleteFileUpload(ctx, f.root, record.ID)
	require(t, err)
	if ready.Status != "ready" || len(ready.SHA256) != 64 {
		t.Fatal("upload not verified")
	}
	_, err = f.b.CompleteFileUpload(ctx, f.root, record.ID)
	require(t, err)
	role := f.role(t, "files_reader", resources.FileRead, resources.FileUpload)
	var otherAdmin model.Administrator
	require(t, f.db.Where("username = ?", "file_other").First(&otherAdmin).Error)
	f.assign(t, otherAdmin.ID, role)
	_, err = f.a.DownloadFile(ctx, other, record.ID)
	code(t, err, codes.NotFound)
	list, count, err := f.a.ListFiles(ctx, other, FileTaskQuery{Page: 1, PageSize: 20})
	require(t, err)
	if count != 0 || len(list) != 0 {
		t.Fatal("private metadata leaked")
	}
	bad, err := f.a.CreateFile(ctx, f.root, "fake.png", "image/png", 5)
	require(t, err)
	require(t, objects.Put(ctx, bad.ObjectKey, strings.NewReader("hello"), 5))
	_, err = f.a.CompleteFileUpload(ctx, f.root, bad.ID)
	code(t, err, codes.InvalidArgument)
	require(t, f.a.DeleteFile(ctx, f.root, record.ID))
	_, err = f.a.DownloadFile(ctx, f.root, record.ID)
	code(t, err, codes.FailedPrecondition)
	var job model.Task
	require(t, f.db.Where("file_id = ?", record.ID).First(&job).Error)
	require(t, worker.Process(ctx, task.Payload{ID: job.ID, Generation: job.Generation}))
	require(t, worker.Process(ctx, task.Payload{ID: job.ID, Generation: job.Generation}))
	if _, err = objects.Open(ctx, record.ObjectKey); err == nil {
		t.Fatal("deleted object remains")
	}
	var audits int64
	require(t, f.db.Model(&model.AuditLog{}).Where("target_type = 'file' AND target_id = ?", record.ID).Count(&audits).Error)
	if audits != 2 {
		t.Fatalf("want one upload and one delete audit; got %d", audits)
	}
}
func TestExportOutboxIdempotencyCrashAndPermissions(t *testing.T) {
	f := setup(t)
	objects, worker := fileJobs(t, f)
	_, err := f.a.CreateDictionary(ctx, f.root, "export.test", DictionaryInput{Name: "=1+1", Enabled: true, Items: []model.DictionaryItem{{Value: "1", Label: "测试", Enabled: true}}})
	require(t, err)
	request := "6fd751c2-44ab-4c6d-8d04-80b2e4e0dc59"
	var ids [2]uint64
	var wg sync.WaitGroup
	for i := range ids {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			job, err := f.a.ExportDictionaries(ctx, f.root, request)
			if err != nil {
				t.Error(err)
				return
			}
			ids[i] = job.ID
		}(i)
	}
	wg.Wait()
	if ids[0] == 0 || ids[0] != ids[1] {
		t.Fatal("request idempotency broken")
	}
	err = worker.Dispatch(ctx, func(context.Context, *asynq.Task, ...asynq.Option) (*asynq.TaskInfo, error) {
		return nil, errors.New("queue unavailable")
	}, "goseed_test_system")
	if err == nil {
		t.Fatal("queue error ignored")
	}
	var outbox int64
	require(t, f.db.Model(&model.TaskOutbox{}).Count(&outbox).Error)
	if outbox != 1 {
		t.Fatal("committed task lost with queue outage")
	}
	// A worker dies after claiming: an expired lease must be recovered by another attempt.
	require(t, f.db.Model(&model.Task{}).Where("id = ?", ids[0]).Updates(map[string]any{"status": "running", "attempts": 1, "lease_until": time.Now().Add(-time.Minute)}).Error)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := worker.Process(ctx, task.Payload{ID: ids[0], Generation: 1}); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	var job model.Task
	require(t, f.db.First(&job, ids[0]).Error)
	if job.Status != "succeeded" || job.Attempts != 2 {
		t.Fatalf("bad recovered task: %+v", job)
	}
	file, err := f.a.DownloadFile(ctx, f.root, job.FileID)
	require(t, err)
	if !strings.HasPrefix(file.ObjectKey, file.CreatedAt.UTC().Format("20060102/")) || len(file.ObjectKey) != 41 {
		t.Fatalf("export key does not use its creation date: %q", file.ObjectKey)
	}
	r, err := objects.Open(ctx, file.ObjectKey)
	require(t, err)
	data, err := io.ReadAll(r)
	r.Close()
	require(t, err)
	if !strings.Contains(string(data), "'=1+1") {
		t.Fatal("CSV injection not escaped")
	}
	require(t, worker.Process(ctx, task.Payload{ID: job.ID, Generation: 1}))
	require(t, f.db.First(&job, job.ID).Error)
	if job.Attempts != 2 {
		t.Fatal("completed task reran")
	}
	user, token := f.user(t, "exporter")
	role := f.role(t, "export_role", resources.TaskRead, resources.FileRead, resources.DictionaryRead, resources.TaskExport, resources.TaskRetry)
	f.assign(t, user, role)
	denied, err := f.a.ExportDictionaries(ctx, token, "6fd751c2-44ab-4c6d-8d04-80b2e4e0dc60")
	require(t, err)
	require(t, f.a.GrantRole(ctx, f.root, role, []string{resources.TaskRead, resources.TaskRetry}, f.roleRevision(t, role)))
	require(t, worker.Process(ctx, task.Payload{ID: denied.ID, Generation: 1}))
	require(t, f.db.First(denied, denied.ID).Error)
	if denied.Status != "failed" || !strings.Contains(denied.Error, "权限") {
		t.Fatal("revoked export still ran")
	}
	code(t, f.a.RetryTask(ctx, token, denied.ID), codes.PermissionDenied)
	require(t, f.a.GrantRole(ctx, f.root, role, []string{resources.TaskRead, resources.FileRead, resources.DictionaryRead, resources.TaskExport, resources.TaskRetry}, f.roleRevision(t, role)))
	// Failed export artifacts may be cleaned; a later authorized retry gets a new immutable key.
	var failedFile model.File
	require(t, f.db.First(&failedFile, denied.FileID).Error)
	oldKey := failedFile.ObjectKey
	require(t, f.db.Model(&failedFile).Updates(map[string]any{"status": "deleted"}).Error)
	require(t, f.a.RetryTask(ctx, token, denied.ID))
	require(t, f.db.First(&failedFile, denied.FileID).Error)
	if failedFile.Status != "pending" || failedFile.ObjectKey == oldKey {
		t.Fatal("retry did not replace cleaned artifact key")
	}
	if !strings.HasPrefix(failedFile.ObjectKey, failedFile.CreatedAt.UTC().Format("20060102/")) {
		t.Fatal("retried export key does not use its new creation date")
	}

	require(t, worker.Process(ctx, task.Payload{ID: denied.ID, Generation: 1}))
	require(t, f.db.First(denied, denied.ID).Error)
	if denied.Status != "pending" || denied.Generation != 2 {
		t.Fatal("stale queue message changed retried task")
	}
	require(t, worker.Process(ctx, task.Payload{ID: denied.ID, Generation: 2}))
	require(t, f.db.First(denied, denied.ID).Error)
	if denied.Status != "succeeded" {
		t.Fatal("manual retry failed")
	}
}
func TestTaskAuditRollbackAndHourlyCleanup(t *testing.T) {
	f := setup(t)
	objects, worker := fileJobs(t, f)
	callback := "test:file-task-audit"
	require(t, f.db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "audit_logs" {
			tx.AddError(errors.New("audit unavailable"))
		}
	}))
	_, err := f.a.ExportDictionaries(ctx, f.root, "6fd751c2-44ab-4c6d-8d04-80b2e4e0dc61")
	code(t, err, codes.Unavailable)
	require(t, f.db.Callback().Create().Remove(callback))
	for _, table := range []string{"files", "tasks", "task_outbox"} {
		var n int64
		require(t, f.db.Table(table).Count(&n).Error)
		if n != 0 {
			t.Fatalf("%s escaped rolled-back audit", table)
		}
	}
	expired, err := f.a.CreateFile(ctx, f.root, "abandoned.txt", "text/plain", 5)
	require(t, err)
	require(t, objects.Put(ctx, expired.ObjectKey, strings.NewReader("hello"), 5))
	require(t, f.db.Model(expired).Update("created_at", time.Now().Add(-25*time.Hour)).Error)
	failed, err := f.a.ExportDictionaries(ctx, f.root, "6fd751c2-44ab-4c6d-8d04-80b2e4e0dc63")
	require(t, err)
	require(t, f.db.Model(failed).Update("status", "failed").Error)
	require(t, f.db.Model(&model.File{}).Where("id = ?", failed.FileID).Update("created_at", time.Now().Add(-25*time.Hour)).Error)
	tick := task.TickMessage(time.Now())
	require(t, worker.Schedule(ctx, tick.Payload()))
	require(t, worker.Schedule(ctx, tick.Payload()))
	var scheduled []model.Task
	require(t, f.db.Where("kind = ?", task.Cleanup).Find(&scheduled).Error)
	if len(scheduled) != 1 {
		t.Fatal("duplicate schedule rows")
	}
	require(t, worker.Process(ctx, task.Payload{ID: scheduled[0].ID, Generation: 1}))
	require(t, f.db.First(expired, expired.ID).Error)
	if expired.Status != "deleted" {
		t.Fatal("expired upload not cleaned")
	}
	var failedFile model.File
	require(t, f.db.First(&failedFile, failed.FileID).Error)
	if failedFile.Status != "deleted" {
		t.Fatal("failed export residue not cleaned")
	}

}
func TestAsynqDeliveryThroughWorker(t *testing.T) {
	f := setup(t)
	_, worker := fileJobs(t, f)
	dbIndex, err := strconv.Atoi(os.Getenv("QUEUE_DB"))
	require(t, err)
	opt := asynq.RedisClientOpt{Addr: os.Getenv("QUEUE_HOST") + ":" + os.Getenv("QUEUE_PORT"), Password: os.Getenv("QUEUE_PASSWORD"), DB: dbIndex}
	client := asynq.NewClient(opt)
	defer client.Close()
	require(t, client.Ping())
	queue := fmt.Sprintf("goseed_test_%d_system", time.Now().UnixNano())
	inspector := asynq.NewInspector(opt)
	defer inspector.Close()
	defer func() {
		if err := inspector.DeleteQueue(queue, true); err != nil && !errors.Is(err, asynq.ErrQueueNotFound) {
			t.Error(err)
		}
	}()
	server := asynq.NewServer(opt, asynq.Config{Concurrency: 2, Queues: map[string]int{queue: 1}, ShutdownTimeout: time.Second, TaskCheckInterval: 100 * time.Millisecond})
	mux := asynq.NewServeMux()
	mux.HandleFunc(task.Execute, func(ctx context.Context, msg *asynq.Task) error {
		p, err := task.Decode(msg.Payload())
		if err != nil {
			return err
		}
		return worker.Process(ctx, p)
	})
	require(t, server.Start(mux))
	defer server.Shutdown()
	job, err := f.a.ExportDictionaries(ctx, f.root, "6fd751c2-44ab-4c6d-8d04-80b2e4e0dc62")
	require(t, err)
	require(t, worker.Dispatch(ctx, client.EnqueueContext, queue))
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		require(t, f.db.First(job, job.ID).Error)
		if job.Status == "succeeded" {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("Asynq did not complete export: %s", job.Status)
}
