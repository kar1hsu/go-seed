// Package jobs runs recoverable system use cases shared by workers and integration tests.
package jobs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/kar1hsu/go-seed/pkg/storage"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/internal/authorization"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/internal/task"
	"github.com/kar1hsu/go-seed/services/system/resources"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Service struct {
	store         *repository.FilesTasks
	rbac          *repository.RBAC
	auth          *authorization.Authorizer
	objects       storage.Store
	retentionDays int
}

func New(store *repository.FilesTasks, rbac *repository.RBAC, auth *authorization.Authorizer, objects storage.Store) *Service {
	return &Service{store: store, rbac: rbac, auth: auth, objects: objects}
}

// SetAuditRetention is applied once during explicit process assembly; 0 retains all logs.
func (s *Service) SetAuditRetention(days int) { s.retentionDays = days }

// Dispatch keeps the outbox until business completion. Periodic redelivery also repairs lost Redis data.
func (s *Service) Dispatch(ctx context.Context, enqueue func(context.Context, *asynq.Task, ...asynq.Option) (*asynq.TaskInfo, error), queue string) error {
	now := time.Now().UTC()
	rows, err := s.store.PendingOutbox(ctx, now)
	if err != nil {
		return err
	}
	for _, o := range rows {
		job, err := s.store.Task(ctx, o.TaskID, false)
		if err != nil {
			return err
		}
		if job.Status == "succeeded" || job.Status == "failed" || job.Generation != o.Generation {
			continue
		}
		reserved, err := s.store.ReserveDelivery(ctx, o, now)
		if err != nil {
			return err
		}
		if !reserved {
			continue
		}
		id := fmt.Sprintf("job-%d-%d-delivery-%d", job.ID, job.Generation, o.Delivery+1)
		_, err = enqueue(ctx, task.Message(job.ID, job.Generation), asynq.Queue(queue), asynq.TaskID(id), asynq.Timeout(time.Minute), asynq.MaxRetry(10), asynq.Retention(2*time.Minute))
		if err != nil && !errors.Is(err, asynq.ErrTaskIDConflict) {
			return err
		}
		if err = s.store.Delivered(ctx, o, now); err != nil {
			return err
		}
	}
	return nil
}
func (s *Service) Schedule(ctx context.Context, data []byte) error {
	var p task.Schedule
	if err := json.Unmarshal(data, &p); err != nil || p.Slot <= 0 || p.Slot%3600 != 0 {
		return asynq.SkipRetry
	}
	if p.Slot > time.Now().UTC().Truncate(time.Hour).Unix() {
		return errors.New("schedule slot is ahead of the worker clock")
	}
	key := fmt.Sprintf("upload-cleanup:%d", p.Slot)
	return s.rbac.Transaction(ctx, true, func(tx *repository.RBAC) error {
		_, err := tx.FilesTasks().TaskByKey(ctx, key)
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		now := time.Now().UTC()
		return tx.FilesTasks().CreateTask(ctx, &model.Task{Kind: task.Cleanup, DedupKey: key, Status: "pending", Generation: 1, MaxAttempts: 5, CreatedAt: now, UpdatedAt: now, NextRunAt: now})
	})
}
func (s *Service) Process(ctx context.Context, p task.Payload) error {
	var job *model.Task
	err := s.store.Transaction(ctx, func(tx *repository.FilesTasks) error {
		t, err := tx.Task(ctx, p.ID, true)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if t.Generation != p.Generation || t.Status == "succeeded" || t.Status == "failed" || t.NextRunAt.After(now) || (t.Status == "running" && t.LeaseUntil != nil && t.LeaseUntil.After(now)) {
			return nil
		}
		if t.Attempts >= t.MaxAttempts {
			t.Status = "failed"
			t.Error = "任务多次中断，请人工重试"
			t.UpdatedAt = now
			t.LeaseUntil = nil
			if err = tx.SaveTask(ctx, t); err != nil {
				return err
			}
			return tx.DropOutbox(ctx, t.ID)
		}
		t.Attempts++
		t.Status = "running"
		t.UpdatedAt = now
		lease := now.Add(2 * time.Minute)
		t.LeaseUntil = &lease
		if err = tx.SaveTask(ctx, t); err != nil {
			return err
		}
		job = t
		return nil
	})
	if err != nil || job == nil {
		return err
	}
	if id, err := trace.TraceIDFromHex(job.TraceID); err == nil {
		ctx = trace.ContextWithSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{TraceID: id, Remote: true}))
	}
	runCtx, cancel := context.WithTimeout(ctx, 50*time.Second)
	defer cancel()
	var size int64
	var sum string
	switch job.Kind {
	case task.Export:
		size, sum, err = s.export(runCtx, job)
	case task.Delete:
		err = s.deleteFile(runCtx, job.FileID)
	case task.Cleanup:
		err = s.cleanup(runCtx)
	default:
		err = xerr.New(codes.InvalidArgument)
	}
	if err == nil {
		err = s.complete(runCtx, job, size, sum)
	}
	if err == nil {
		logx.WithContext(ctx).Infow("system task completed", logx.Field("taskId", job.ID), logx.Field("attempt", job.Attempts))
		return nil
	}
	logx.WithContext(ctx).Errorw("system task attempt failed", logx.Field("taskId", job.ID), logx.Field("attempt", job.Attempts), logx.Field("errorType", fmt.Sprintf("%T", err)))
	// Parent may have been canceled during shutdown. A short independent write records failure;
	// if this also fails, the durable lease/outbox recovers after expiry.
	saveCtx, stop := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	defer stop()
	return s.failed(saveCtx, job, err)
}
func (s *Service) authorized(ctx context.Context, tx *repository.RBAC, owner uint64) error {
	a, err := s.auth.Deferred(ctx, tx, owner)
	if err != nil {
		return err
	}
	if !a.Covers([]string{resources.TaskExport, resources.DictionaryRead, resources.FileRead}) {
		return xerr.New(codes.PermissionDenied)
	}
	return nil
}

// Neutralize formula prefixes before encoding cells for spreadsheet applications.
func CSVCell(v string) string {
	trim := strings.TrimLeft(v, " \t\r\n")
	if trim != "" && strings.ContainsAny(trim[:1], "=+-@") || strings.HasPrefix(v, "\t") || strings.HasPrefix(v, "\r") {
		return "'" + v
	}
	return v
}
func (s *Service) export(ctx context.Context, job *model.Task) (int64, string, error) {
	var dictionaries []model.Dictionary
	var items []model.DictionaryItem
	err := s.rbac.Transaction(ctx, false, func(tx *repository.RBAC) error {
		if err := s.authorized(ctx, tx, job.OwnerID); err != nil {
			return err
		}
		var err error
		dictionaries, items, err = tx.FilesTasks().DictionaryExport(ctx)
		return err
	})
	if err != nil {
		return 0, "", err
	}
	if len(dictionaries) > 1000 || len(items) > 10000 {
		return 0, "", xerr.New(codes.ResourceExhausted)
	}
	var data bytes.Buffer
	data.WriteString("\xef\xbb\xbf")
	w := csv.NewWriter(&data)
	w.Write([]string{"字典标识", "字典名称", "字典状态", "字典项值", "字典项标签", "排序", "字典项状态"})
	groups := map[uint64][]model.DictionaryItem{}
	for _, i := range items {
		groups[i.DictionaryID] = append(groups[i.DictionaryID], i)
	}
	for _, d := range dictionaries {
		children := groups[d.ID]
		if len(children) == 0 {
			children = []model.DictionaryItem{{}}
		}
		for _, i := range children {
			w.Write([]string{CSVCell(d.Code), CSVCell(d.Name), strconv.FormatBool(d.Enabled), CSVCell(i.Value), CSVCell(i.Label), strconv.FormatInt(i.Sort, 10), strconv.FormatBool(i.Enabled)})
		}
	}
	w.Flush()
	if err = w.Error(); err != nil {
		return 0, "", err
	}
	if data.Len() > int(storage.MaxSize) {
		return 0, "", xerr.New(codes.ResourceExhausted)
	}
	f, err := s.store.File(ctx, job.FileID)
	if err != nil {
		return 0, "", err
	}
	if f.Status != "pending" {
		return 0, "", xerr.New(codes.FailedPrecondition)
	}
	err = s.objects.Put(ctx, f.ObjectKey, bytes.NewReader(data.Bytes()), int64(data.Len()))
	if err != nil && !errors.Is(err, storage.ErrExists) {
		return 0, "", err
	}
	// A crashed attempt may already have written this immutable result; resume the same export.
	r, err := s.objects.Open(ctx, f.ObjectKey)
	if err != nil {
		return 0, "", err
	}
	defer r.Close()
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(r, storage.MaxSize+1))
	if err != nil {
		return 0, "", err
	}
	if n > storage.MaxSize {
		return 0, "", xerr.New(codes.ResourceExhausted)
	}
	return n, hex.EncodeToString(h.Sum(nil)), nil
}
func (s *Service) complete(ctx context.Context, job *model.Task, size int64, sum string) error {
	return s.rbac.Transaction(ctx, true, func(tx *repository.RBAC) error {
		r := tx.FilesTasks()
		current, err := r.Task(ctx, job.ID, true)
		if err != nil {
			return err
		}
		if !sameAttempt(current, job) {
			return nil
		}
		if job.Kind == task.Export {
			if err = s.authorized(ctx, tx, job.OwnerID); err != nil {
				return err
			}
			f, err := r.File(ctx, job.FileID)
			if err != nil {
				return err
			}
			if f.Status != "pending" {
				return xerr.New(codes.FailedPrecondition)
			}
			f.Status, f.Size, f.SHA256, f.UpdatedAt = "ready", size, sum, time.Now().UTC()
			if err = r.SaveFile(ctx, f); err != nil {
				return err
			}
		}
		current.Status = "succeeded"
		current.Error = ""
		current.LeaseUntil = nil
		current.UpdatedAt = time.Now().UTC()
		if err = r.SaveTask(ctx, current); err != nil {
			return err
		}
		return r.DropOutbox(ctx, current.ID)
	})
}
func sameAttempt(a, b *model.Task) bool {
	return a.Generation == b.Generation && a.Attempts == b.Attempts && a.Status == "running"
}
func (s *Service) failed(ctx context.Context, job *model.Task, cause error) error {
	return s.store.Transaction(ctx, func(tx *repository.FilesTasks) error {
		current, err := tx.Task(ctx, job.ID, true)
		if err != nil {
			return err
		}
		if !sameAttempt(current, job) {
			return nil
		}
		permanent := false
		message := "存储或依赖暂时不可用，将自动重试"
		switch status.Code(cause) {
		case codes.PermissionDenied:
			permanent = true
			message = "发起人的账号或导出权限已失效"
		case codes.ResourceExhausted:
			permanent = true
			message = "导出超过 1000 个字典、10000 个字典项或 10 MiB 限制"
		case codes.InvalidArgument, codes.FailedPrecondition, codes.NotFound:
			permanent = true
			message = "任务参数或关联文件状态无效"
		}
		current.Status = "pending"
		current.Error = message
		current.LeaseUntil = nil
		current.UpdatedAt = time.Now().UTC()
		current.NextRunAt = current.UpdatedAt.Add(time.Duration(1<<min(current.Attempts, 6)) * time.Second)
		if permanent || current.Attempts >= current.MaxAttempts {
			current.Status = "failed"
			if !permanent {
				current.Error = "多次重试后依赖仍不可用，请检查服务后重试"
			}
		}
		if err = tx.SaveTask(ctx, current); err != nil {
			return err
		}
		if current.Status == "failed" {
			return tx.DropOutbox(ctx, current.ID)
		}
		return tx.PutOutbox(ctx, current)
	})
}
func (s *Service) deleteFile(ctx context.Context, id uint64) error {
	f, err := s.store.File(ctx, id)
	if err != nil {
		return err
	}
	if f.Status == "deleted" {
		return nil
	}
	if f.Status != "deleting" {
		return xerr.New(codes.FailedPrecondition)
	}
	if err = s.objects.Delete(ctx, f.ObjectKey); err != nil {
		return err
	}
	return s.store.MarkDeleted(ctx, id)
}
func (s *Service) cleanup(ctx context.Context) error {
	cutoff := time.Now().UTC().Add(-24 * time.Hour)
	// Bound a run; remaining rows are handled by the next hourly schedule.
	rows, err := s.store.ExpiredUploads(ctx, cutoff)
	if err != nil {
		return err
	}
	var failures []error
	for _, f := range rows {
		if err = s.rbac.Transaction(ctx, true, func(tx *repository.RBAC) error { return tx.FilesTasks().MarkDeleting(ctx, f.ID, cutoff) }); err != nil {
			return err
		}
		current, err := s.store.File(ctx, f.ID)
		if err != nil {
			return err
		}
		if current.Status == "deleting" {
			if err = s.deleteFile(ctx, f.ID); err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				// Do not let a permanently broken object monopolize the first batch.
				if deferErr := s.store.DeferCleanup(ctx, f.ID, time.Now().UTC().Add(time.Hour)); deferErr != nil {
					return deferErr
				}
				logx.WithContext(ctx).Errorw("file cleanup deferred", logx.Field("fileId", f.ID), logx.Field("errorType", fmt.Sprintf("%T", err)))
				failures = append(failures, fmt.Errorf("file %d cleanup failed", f.ID))
			}
		}
	}
	if s.retentionDays > 0 {
		failures = append(failures, s.rbac.Audit().Prune(ctx, time.Now().UTC().AddDate(0, 0, -s.retentionDays)))
	}
	return errors.Join(append(failures, s.objects.CleanupTemporary(ctx, cutoff))...)
}
