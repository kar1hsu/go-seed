//go:build integration

package management

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/kar1hsu/go-seed/pkg/storage"
	"github.com/kar1hsu/go-seed/services/system/internal/application/jobs"
	"github.com/kar1hsu/go-seed/services/system/internal/filemeta"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/internal/task"
	"github.com/kar1hsu/go-seed/services/system/resources"
)

func TestLongFilenameAuditAndDelete(t *testing.T) {
	f := setup(t)
	objects, _ := fileJobs(t, f)
	name := strings.Repeat("长", 176) + ".txt"
	record, err := f.a.CreateFile(ctx, f.root, name, "text/plain", 5)
	require(t, err)
	require(t, objects.Put(ctx, record.ObjectKey, strings.NewReader("hello"), 5))
	_, err = f.a.CompleteFileUpload(ctx, f.root, record.ID)
	require(t, err)
	require(t, f.a.DeleteFile(ctx, f.root, record.ID))
	var audit []model.AuditLog
	require(t, f.db.Where("target_type = 'file' AND target_id = ?", record.ID).Find(&audit).Error)
	if len(audit) != 2 {
		t.Fatalf("want both audit events, got %d", len(audit))
	}
	for _, a := range audit {
		if a.TargetName != name {
			t.Fatal("filename audit truncated")
		}
	}
}

func TestOutboxRedeliveryUsesFreshIDWithoutBusinessAttempt(t *testing.T) {
	f := setup(t)
	_, worker := fileJobs(t, f)
	job, err := f.a.ExportDictionaries(ctx, f.root, "6fd751c2-44ab-4c6d-8d04-80b2e4e0dc78")
	require(t, err)
	seen := map[string]bool{}
	enqueue := func(_ context.Context, _ *asynq.Task, options ...asynq.Option) (*asynq.TaskInfo, error) {
		var id string
		for _, o := range options {
			if o.Type() == asynq.TaskIDOpt {
				id = o.Value().(string)
			}
		}
		if id == "" || seen[id] {
			t.Fatal("redelivery reused an archived transport ID")
		}
		seen[id] = true
		// An uncertain enqueue result can already have created/archived this ID.
		return nil, errors.New("lost enqueue response")
	}
	for range 2 {
		if worker.Dispatch(ctx, enqueue, "test") == nil {
			t.Fatal("enqueue failure lost")
		}
		require(t, f.db.Model(&model.TaskOutbox{}).Where("task_id = ?", job.ID).Update("available_at", time.Now().Add(-time.Second)).Error)
	}
	var current model.Task
	require(t, f.db.First(&current, job.ID).Error)
	if len(seen) != 2 || current.Attempts != 0 {
		t.Fatal("transport and business attempts coupled")
	}
	// Competing dispatchers must not reserve the same delivery snapshot.
	store := repository.NewFilesTasks(f.db)
	rows, err := store.PendingOutbox(ctx, time.Now())
	require(t, err)
	ok, err := store.ReserveDelivery(ctx, rows[0], time.Now())
	require(t, err)
	if !ok {
		t.Fatal("reservation failed")
	}
	ok, err = store.ReserveDelivery(ctx, rows[0], time.Now())
	require(t, err)
	if ok {
		t.Fatal("stale dispatcher reserved twice")
	}
}

func TestRetainDisabledRoleWithoutGrantingNewDisabledRole(t *testing.T) {
	f := setup(t)
	id, _ := f.user(t, "retained_role_user")
	retained := f.role(t, "retained_role", resources.UserRead)
	added := f.role(t, "added_role", resources.RoleRead)
	f.assign(t, id, retained)
	require(t, f.a.UpdateRole(ctx, f.root, retained, "retained", "", false, f.roleRevision(t, retained)))
	require(t, f.a.AssignRoles(ctx, f.root, id, []uint64{retained, added}, f.adminRevision(t, id)))
	require(t, f.a.AssignRoles(ctx, f.root, id, []uint64{added}, f.adminRevision(t, id)))
	if f.a.AssignRoles(ctx, f.root, id, []uint64{retained, added}, f.adminRevision(t, id)) == nil {
		t.Fatal("new disabled assignment allowed")
	}
}

type brokenDelete struct {
	storage.Store
	key string
}

func (s brokenDelete) Delete(ctx context.Context, key string) error {
	if key == s.key {
		return errors.New("object deletion denied")
	}
	return s.Store.Delete(ctx, key)
}
func TestCleanupContinuesPastBrokenObject(t *testing.T) {
	f := setup(t)
	objects, _ := fileJobs(t, f)
	store := repository.NewFilesTasks(f.db)
	bad := model.File{Name: "bad.txt", ObjectKey: filemeta.Key(time.Now(), time.UTC), Status: "deleting", Source: "upload", CreatedAt: time.Now().Add(-48 * time.Hour), UpdatedAt: time.Now()}
	good := model.File{Name: "good.txt", ObjectKey: filemeta.Key(time.Now(), time.UTC), Status: "deleting", Source: "upload", CreatedAt: bad.CreatedAt, UpdatedAt: bad.UpdatedAt}
	require(t, store.CreateFile(ctx, &bad))
	require(t, store.CreateFile(ctx, &good))
	worker := jobs.New(store, repository.NewRBAC(f.db), f.a.authorizer, brokenDelete{objects, bad.ObjectKey})
	job := newTask(ctx, 0, task.Cleanup, "cleanup-fixture", 0)
	require(t, store.CreateTask(ctx, job))
	require(t, worker.Process(ctx, task.Payload{ID: job.ID, Generation: job.Generation}))
	a, err := store.File(ctx, bad.ID)
	require(t, err)
	b, err := store.File(ctx, good.ID)
	require(t, err)
	if a.Status != "deleting" || a.CleanupAfter == nil || !a.CleanupAfter.After(time.Now()) || b.Status != "deleted" {
		t.Fatal("broken file blocked later cleanup or was forgotten")
	}
	rows, err := store.ExpiredUploads(ctx, time.Now().Add(-24*time.Hour))
	require(t, err)
	if len(rows) != 0 {
		t.Fatal("deferred file monopolizes next batch")
	}
}

func TestLiteralLikeAndNoOpPolicy(t *testing.T) {
	f := setup(t)
	percent, err := f.a.CreateAdministrator(ctx, f.root, "literal_percent", "Name%One", "long-enough-password")
	require(t, err)
	_, err = f.a.CreateAdministrator(ctx, f.root, "literal_plain", "NameXOne", "long-enough-password")
	require(t, err)
	rows, total, err := f.a.ListAdministrators(ctx, f.root, Query{Page: 1, PageSize: 20, Keyword: "%", Enabled: "all"})
	require(t, err)
	if total != 1 || rows[0].ID != percent.ID {
		t.Fatal("LIKE percent interpreted as wildcard")
	}
	role, err := f.a.CreateRole(ctx, f.root, "literal_role", "Role_One", "")
	require(t, err)
	_, err = f.a.CreateRole(ctx, f.root, "literal_other", "RoleXOne", "")
	require(t, err)
	roles, total, err := f.a.ListRoles(ctx, f.root, Query{Page: 1, PageSize: 20, Keyword: "Role_", Enabled: "all"})
	require(t, err)
	if total != 1 || roles[0].ID != role.ID {
		t.Fatal("LIKE underscore interpreted as wildcard")
	}
	version, err := repository.NewRBAC(f.db).Version(ctx)
	require(t, err)
	var count int64
	require(t, f.db.Model(&model.AuditLog{}).Count(&count).Error)
	require(t, f.a.UpdateAdministrator(ctx, f.root, percent.ID, percent.DisplayName, true, Revision(percent.UpdatedAt)))
	require(t, f.a.AssignRoles(ctx, f.root, percent.ID, nil, Revision(percent.UpdatedAt)))
	require(t, f.a.UpdateRole(ctx, f.root, role.ID, role.Name, role.Description, true, Revision(role.UpdatedAt)))
	require(t, f.a.GrantRole(ctx, f.root, role.ID, nil, Revision(role.UpdatedAt)))
	after, err := repository.NewRBAC(f.db).Version(ctx)
	require(t, err)
	var newCount int64
	require(t, f.db.Model(&model.AuditLog{}).Count(&newCount).Error)
	if version != after || count != newCount {
		t.Fatal("no-op mutated audit or policy version")
	}
}
func TestAuditRetentionIsBoundedAndOptIn(t *testing.T) {
	f := setup(t)
	objects, worker := fileJobs(t, f)
	old := model.AuditLog{Action: "test", TargetType: "test", ChangesJSON: "[]", CreatedAt: time.Now().AddDate(-2, 0, 0)}
	require(t, f.db.Create(&old).Error)
	// The default service retains history, even during cleanup.
	run := func(service *jobs.Service, key string) {
		job := newTask(ctx, 0, task.Cleanup, key, 0)
		require(t, repository.NewFilesTasks(f.db).CreateTask(ctx, job))
		require(t, service.Process(ctx, task.Payload{ID: job.ID, Generation: job.Generation}))
	}
	run(worker, "retain-default")
	var count int64
	require(t, f.db.Model(&model.AuditLog{}).Where("id = ?", old.ID).Count(&count).Error)
	if count != 1 {
		t.Fatal("default retention deleted history")
	}
	next := jobs.New(repository.NewFilesTasks(f.db), repository.NewRBAC(f.db), f.a.authorizer, objects)
	next.SetAuditRetention(365)
	run(next, "retain-configured")
	require(t, f.db.Model(&model.AuditLog{}).Where("id = ?", old.ID).Count(&count).Error)
	if count != 0 {
		t.Fatal("configured retention not applied")
	}
}
