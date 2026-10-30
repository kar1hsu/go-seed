package repository

import (
	"context"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

type FilesTasks struct{ db *gorm.DB }

func (r *RBAC) FilesTasks() *FilesTasks     { return &FilesTasks{r.db} }
func NewFilesTasks(db *gorm.DB) *FilesTasks { return &FilesTasks{db} }
func (r *FilesTasks) Transaction(ctx context.Context, fn func(*FilesTasks) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return fn(&FilesTasks{tx}) })
}
func (r *FilesTasks) File(ctx context.Context, id uint64) (*model.File, error) {
	var f model.File
	err := r.db.WithContext(ctx).First(&f, id).Error
	return &f, err
}
func (r *FilesTasks) CreateFile(ctx context.Context, f *model.File) error {
	return r.db.WithContext(ctx).Create(f).Error
}
func (r *FilesTasks) SaveFile(ctx context.Context, f *model.File) error {
	return r.db.WithContext(ctx).Save(f).Error
}
func (r *FilesTasks) ListFiles(ctx context.Context, owner uint64, page, size int64, keyword, status string) ([]model.File, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.File{}).Where("owner_id = ? AND status <> 'deleted'", owner)
	if keyword != "" {
		q = q.Where("name LIKE ? ESCAPE '!'", likePattern(keyword))
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := []model.File{}
	err := q.Order("id DESC").Offset(int((page - 1) * size)).Limit(int(size)).Find(&items).Error
	return items, total, err
}
func likePattern(s string) string {
	return "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(s) + "%"
}
func (r *FilesTasks) Task(ctx context.Context, id uint64, lock bool) (*model.Task, error) {
	q := r.db.WithContext(ctx)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var t model.Task
	err := q.First(&t, id).Error
	return &t, err
}
func (r *FilesTasks) TaskByKey(ctx context.Context, key string) (*model.Task, error) {
	var t model.Task
	err := r.db.WithContext(ctx).Where("dedup_key = ?", key).First(&t).Error
	return &t, err
}
func (r *FilesTasks) CreateTask(ctx context.Context, t *model.Task) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return err
	}
	return r.PutOutbox(ctx, t)
}
func (r *FilesTasks) SaveTask(ctx context.Context, t *model.Task) error {
	return r.db.WithContext(ctx).Save(t).Error
}
func (r *FilesTasks) PutOutbox(ctx context.Context, t *model.Task) error {
	o := model.TaskOutbox{TaskID: t.ID, Generation: t.Generation, AvailableAt: t.NextRunAt}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"generation", "available_at"})}).Create(&o).Error
}
func (r *FilesTasks) DropOutbox(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("task_id = ?", id).Delete(&model.TaskOutbox{}).Error
}
func (r *FilesTasks) ListTasks(ctx context.Context, owner uint64, page, size int64, keyword, status string) ([]model.Task, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Task{}).Where("owner_id IN ?", []uint64{owner, 0})
	if keyword != "" {
		q = q.Where("kind LIKE ? ESCAPE '!' OR trace_id = ?", likePattern(keyword), keyword)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := []model.Task{}
	err := q.Order("id DESC").Offset(int((page - 1) * size)).Limit(int(size)).Find(&items).Error
	return items, total, err
}
func (r *FilesTasks) PendingOutbox(ctx context.Context, now time.Time) ([]model.TaskOutbox, error) {
	items := []model.TaskOutbox{}
	err := r.db.WithContext(ctx).Where("available_at <= ?", now).Order("available_at ASC").Limit(50).Find(&items).Error
	return items, err
}

// ReserveDelivery allocates a fresh transport ID before enqueueing. A crash or an
// uncertain Redis result becomes eligible again after the short delivery lease.
func (r *FilesTasks) ReserveDelivery(ctx context.Context, o model.TaskOutbox, now time.Time) (bool, error) {
	result := r.db.WithContext(ctx).Model(&model.TaskOutbox{}).Where("task_id = ? AND generation = ? AND delivery = ? AND available_at = ?", o.TaskID, o.Generation, o.Delivery, o.AvailableAt).Updates(map[string]any{"delivery": o.Delivery + 1, "available_at": now.Add(30 * time.Second)})
	return result.RowsAffected == 1, result.Error
}
func (r *FilesTasks) Delivered(ctx context.Context, o model.TaskOutbox, now time.Time) error {
	return r.db.WithContext(ctx).Model(&model.TaskOutbox{}).Where("task_id = ? AND generation = ? AND delivery = ?", o.TaskID, o.Generation, o.Delivery+1).Update("delivered_at", now).Error
}
func (r *FilesTasks) ExpiredUploads(ctx context.Context, cutoff time.Time) ([]model.File, error) {
	items := []model.File{}
	err := r.db.WithContext(ctx).Where("((status = 'pending' AND created_at < ? AND (source = 'upload' OR EXISTS (SELECT 1 FROM tasks WHERE tasks.file_id = files.id AND tasks.status = 'failed'))) OR status = 'deleting') AND (cleanup_after IS NULL OR cleanup_after <= ?)", cutoff, time.Now().UTC()).Order("id ASC").Limit(200).Find(&items).Error
	return items, err
}
func (r *FilesTasks) MarkDeleting(ctx context.Context, id uint64, cutoff time.Time) error {
	return r.db.WithContext(ctx).Model(&model.File{}).Where("id = ? AND status = 'pending' AND created_at < ? AND (source = 'upload' OR EXISTS (SELECT 1 FROM tasks WHERE tasks.file_id = files.id AND tasks.status = 'failed'))", id, cutoff).Updates(map[string]any{"status": "deleting", "updated_at": time.Now().UTC()}).Error
}
func (r *FilesTasks) MarkDeleted(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&model.File{}).Where("id = ? AND status = 'deleting'", id).Updates(map[string]any{"status": "deleted", "updated_at": time.Now().UTC()}).Error
}
func (r *FilesTasks) DeferCleanup(ctx context.Context, id uint64, until time.Time) error {
	return r.db.WithContext(ctx).Model(&model.File{}).Where("id = ? AND status = 'deleting'", id).Update("cleanup_after", until).Error
}

// A bounded, consistent export snapshot. Oversize exports fail rather than silently truncate.
func (r *FilesTasks) DictionaryExport(ctx context.Context) ([]model.Dictionary, []model.DictionaryItem, error) {
	dictionaries := []model.Dictionary{}
	items := []model.DictionaryItem{}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Order("id ASC").Limit(1001).Find(&dictionaries).Error; err != nil {
			return err
		}
		return tx.Order("dictionary_id ASC, sort ASC, value ASC").Limit(10001).Find(&items).Error
	})
	return dictionaries, items, err
}
