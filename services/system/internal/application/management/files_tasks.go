package management

import (
	"context"
	"errors"
	"fmt"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/internal/audit"
	"github.com/kar1hsu/go-seed/services/system/internal/authorization"
	"github.com/kar1hsu/go-seed/services/system/internal/filemeta"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/internal/task"
	"github.com/kar1hsu/go-seed/services/system/resources"
	"github.com/zeromicro/go-zero/core/trace"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
	"regexp"
	"time"
)

type FileTaskQuery struct {
	Page, PageSize  int64
	Keyword, Status string
}

func validFileTaskQuery(q FileTaskQuery, kind string) bool {
	if q.Page < 1 || q.Page > 100000 || q.PageSize < 1 || q.PageSize > 100 || len([]rune(q.Keyword)) > 64 {
		return false
	}
	if q.Status == "" {
		return true
	}
	if kind == "file" {
		return q.Status == "pending" || q.Status == "ready" || q.Status == "deleting"
	}
	return q.Status == "pending" || q.Status == "running" || q.Status == "succeeded" || q.Status == "failed"
}
func (s *Service) ListFiles(ctx context.Context, raw string, q FileTaskQuery) ([]model.File, int64, error) {
	if !validFileTaskQuery(q, "file") {
		return nil, 0, xerr.New(codes.InvalidArgument)
	}
	items := []model.File{}
	var total int64
	err := s.runData(ctx, raw, resources.FileRead, false, func(tx *repository.RBAC, a *authorization.Access) error {
		var err error
		items, total, err = tx.FilesTasks().ListFiles(ctx, a.Admin.ID, q.Page, q.PageSize, q.Keyword, q.Status)
		return err
	})
	return items, total, err
}
func (s *Service) CreateFile(ctx context.Context, raw, name, kind string, size int64) (*model.File, error) {
	if !filemeta.Valid(name, kind, size) {
		return nil, xerr.New(codes.InvalidArgument)
	}
	now := time.Now().UTC()
	f := &model.File{Name: name, ContentType: kind, Size: size, ObjectKey: filemeta.Key(), Status: "pending", Source: "upload", CreatedAt: now, UpdatedAt: now}
	err := s.runData(ctx, raw, resources.FileUpload, true, func(tx *repository.RBAC, a *authorization.Access) error {
		f.OwnerID = a.Admin.ID
		return tx.FilesTasks().CreateFile(ctx, f)
	})
	return f, err
}
func (s *Service) fileAccess(ctx context.Context, raw string, id uint64, upload bool) (*model.File, error) {
	permission := resources.FileRead
	if upload {
		permission = resources.FileUpload
	}
	var f *model.File
	err := s.runData(ctx, raw, permission, false, func(tx *repository.RBAC, a *authorization.Access) error {
		var err error
		f, err = tx.FilesTasks().File(ctx, id)
		if err != nil {
			return err
		}
		if f.OwnerID != a.Admin.ID {
			return xerr.New(codes.NotFound)
		}
		if upload {
			if f.Source != "upload" || (f.Status != "pending" && f.Status != "ready") || time.Since(f.CreatedAt) > time.Hour {
				return xerr.New(codes.FailedPrecondition)
			}
		} else {
			if f.Status != "ready" {
				return xerr.New(codes.FailedPrecondition)
			}
			if f.Source == "export" {
				return a.Require(resources.TaskExport)
			}
		}
		return nil
	})
	return f, err
}
func (s *Service) AuthorizeFileUpload(ctx context.Context, raw string, id uint64) (*model.File, error) {
	return s.fileAccess(ctx, raw, id, true)
}
func (s *Service) DownloadFile(ctx context.Context, raw string, id uint64) (*model.File, error) {
	return s.fileAccess(ctx, raw, id, false)
}
func (s *Service) CompleteFileUpload(ctx context.Context, raw string, id uint64) (*model.File, error) {
	f, err := s.AuthorizeFileUpload(ctx, raw, id)
	if err != nil {
		return nil, err
	}
	if f.Status == "ready" {
		return f, nil
	}
	if s.objects == nil {
		return nil, xerr.New(codes.Unavailable)
	}
	sum, err := filemeta.Inspect(ctx, s.objects, *f)
	if err != nil {
		return nil, publicError(ctx, err)
	}
	err = s.runData(ctx, raw, resources.FileUpload, true, func(tx *repository.RBAC, a *authorization.Access) error {
		current, err := tx.FilesTasks().File(ctx, id)
		if err != nil {
			return err
		}
		if current.OwnerID != a.Admin.ID {
			return xerr.New(codes.NotFound)
		}
		if current.Status == "ready" {
			f = current
			return nil
		}
		if current.Status != "pending" || time.Since(current.CreatedAt) > time.Hour {
			return xerr.New(codes.FailedPrecondition)
		}
		current.Status, current.SHA256, current.UpdatedAt = "ready", sum, time.Now().UTC()
		if err = tx.FilesTasks().SaveFile(ctx, current); err != nil {
			return err
		}
		f = current
		return appendAudit(ctx, tx, a, resources.FileUpload, "file", f.ID, "", f.Name, nil, audit.File(*f))
	})
	return f, err
}
func (s *Service) DeleteFile(ctx context.Context, raw string, id uint64) error {
	return s.runData(ctx, raw, resources.FileDelete, true, func(tx *repository.RBAC, a *authorization.Access) error {
		f, err := tx.FilesTasks().File(ctx, id)
		if err != nil {
			return err
		}
		if f.OwnerID != a.Admin.ID {
			return xerr.New(codes.NotFound)
		}
		if f.Status == "deleting" || f.Status == "deleted" {
			return nil
		}
		if f.Status != "ready" {
			return xerr.New(codes.FailedPrecondition)
		}
		before := audit.File(*f)
		f.Status, f.UpdatedAt = "deleting", time.Now().UTC()
		if err = tx.FilesTasks().SaveFile(ctx, f); err != nil {
			return err
		}
		job := newTask(ctx, a.Admin.ID, task.Delete, fmt.Sprintf("file-delete:%d", id), id)
		if err = tx.FilesTasks().CreateTask(ctx, job); err != nil {
			return err
		}
		return appendAudit(ctx, tx, a, resources.FileDelete, "file", id, "", f.Name, before, audit.File(*f))
	})
}
func newTask(ctx context.Context, owner uint64, kind, key string, file uint64) *model.Task {
	now := time.Now().UTC()
	return &model.Task{OwnerID: owner, Kind: kind, DedupKey: key, Status: "pending", Generation: 1, MaxAttempts: 5, FileID: file, TraceID: trace.TraceIDFromContext(ctx), NextRunAt: now, CreatedAt: now, UpdatedAt: now}
}
func (s *Service) ExportDictionaries(ctx context.Context, raw, key string) (*model.Task, error) {
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(key) {
		return nil, xerr.New(codes.InvalidArgument)
	}
	var job *model.Task
	err := s.runData(ctx, raw, resources.TaskExport, true, func(tx *repository.RBAC, a *authorization.Access) error {
		if !a.Covers([]string{resources.DictionaryRead, resources.FileRead, resources.TaskRead}) {
			return xerr.New(codes.PermissionDenied)
		}
		dedup := fmt.Sprintf("dictionary-export:%d:%s", a.Admin.ID, key)
		existing, err := tx.FilesTasks().TaskByKey(ctx, dedup)
		if err == nil {
			job = existing
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		now := time.Now().UTC()
		f := &model.File{OwnerID: a.Admin.ID, Name: "dictionaries-" + now.Format("20060102-150405") + ".csv", ObjectKey: filemeta.Key(), ContentType: "text/csv", Status: "pending", Source: "export", CreatedAt: now, UpdatedAt: now}
		if err = tx.FilesTasks().CreateFile(ctx, f); err != nil {
			return err
		}
		job = newTask(ctx, a.Admin.ID, task.Export, dedup, f.ID)
		if err = tx.FilesTasks().CreateTask(ctx, job); err != nil {
			return err
		}
		return appendAudit(ctx, tx, a, resources.TaskExport, "task", job.ID, "", "字典 CSV 导出", nil, audit.Task(*job))
	})
	return job, err
}
func (s *Service) ListTasks(ctx context.Context, raw string, q FileTaskQuery) ([]model.Task, int64, error) {
	if !validFileTaskQuery(q, "task") {
		return nil, 0, xerr.New(codes.InvalidArgument)
	}
	items := []model.Task{}
	var total int64
	err := s.runData(ctx, raw, resources.TaskRead, false, func(tx *repository.RBAC, a *authorization.Access) error {
		var err error
		items, total, err = tx.FilesTasks().ListTasks(ctx, a.Admin.ID, q.Page, q.PageSize, q.Keyword, q.Status)
		for i := range items {
			items[i].Retryable = items[i].Status == "failed" && (items[i].OwnerID == a.Admin.ID || items[i].OwnerID == 0 && a.SuperAdmin)
			if items[i].Kind == task.Export {
				items[i].Retryable = items[i].Retryable && a.Covers([]string{resources.TaskExport, resources.DictionaryRead, resources.FileRead})
			}
		}

		return err
	})
	return items, total, err
}
func (s *Service) RetryTask(ctx context.Context, raw string, id uint64) error {
	return s.runData(ctx, raw, resources.TaskRetry, true, func(tx *repository.RBAC, a *authorization.Access) error {
		job, err := tx.FilesTasks().Task(ctx, id, true)
		if err != nil {
			return err
		}
		if job.OwnerID != a.Admin.ID && (job.OwnerID != 0 || !a.SuperAdmin) {
			return xerr.New(codes.NotFound)
		}
		if job.Status != "failed" {
			return xerr.New(codes.FailedPrecondition)
		}
		if job.Kind == task.Export && !a.Covers([]string{resources.TaskExport, resources.DictionaryRead, resources.FileRead}) {
			return xerr.New(codes.PermissionDenied)
		}
		if job.Kind == task.Export {
			f, err := tx.FilesTasks().File(ctx, job.FileID)
			if err != nil {
				return err
			}
			if f.Status == "deleting" {
				return xerr.New(codes.FailedPrecondition)
			}
			if f.Status == "deleted" {
				f.ObjectKey = filemeta.Key()
				f.Status = "pending"
				f.Size = 0
				f.SHA256 = ""
				f.CreatedAt = time.Now().UTC()
				f.UpdatedAt = f.CreatedAt
				if err = tx.FilesTasks().SaveFile(ctx, f); err != nil {
					return err
				}
			}
		}
		before := audit.Task(*job)
		job.Generation++
		job.Attempts = 0
		job.Status = "pending"
		job.Error = ""
		job.LeaseUntil = nil
		job.NextRunAt = time.Now().UTC()
		job.UpdatedAt = job.NextRunAt
		if err = tx.FilesTasks().SaveTask(ctx, job); err != nil {
			return err
		}
		if err = tx.FilesTasks().PutOutbox(ctx, job); err != nil {
			return err
		}
		return appendAudit(ctx, tx, a, resources.TaskRetry, "task", job.ID, "", job.Kind, before, audit.Task(*job))
	})
}
