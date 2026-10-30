package repository

import (
	"context"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"gorm.io/gorm"
	"strings"
	"time"
)

type Audit struct{ db *gorm.DB }
type AuditFilter struct {
	Page, PageSize           int64
	Keyword, Action, TraceID string
	ActorID                  *uint64
	From, To                 *time.Time
}

func (r *RBAC) Audit() *Audit { return &Audit{r.db} }
func (r *Auth) Audit() *Audit { return &Audit{r.db} }
func (r *Audit) Append(ctx context.Context, item *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(item).Error
}
func (r *Audit) Get(ctx context.Context, id uint64) (*model.AuditLog, error) {
	var item model.AuditLog
	err := r.db.WithContext(ctx).First(&item, id).Error
	return &item, err
}
func (r *Audit) List(ctx context.Context, f AuditFilter) ([]model.AuditLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.AuditLog{})
	if f.Keyword != "" {
		keyword := "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(f.Keyword) + "%"
		q = q.Where("actor_username LIKE ? ESCAPE '!' OR actor_name LIKE ? ESCAPE '!' OR target_key LIKE ? ESCAPE '!' OR target_name LIKE ? ESCAPE '!'", keyword, keyword, keyword, keyword)
	}
	if f.Action != "" {
		q = q.Where("action = ?", f.Action)
	}
	if f.ActorID != nil {
		q = q.Where("actor_id = ?", *f.ActorID)
	}
	if f.TraceID != "" {
		q = q.Where("trace_id = ?", f.TraceID)
	}
	if f.From != nil {
		q = q.Where("created_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("created_at < ?", *f.To)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := []model.AuditLog{}
	err := q.Omit("changes").Order("created_at DESC, id DESC").Offset(int((f.Page - 1) * f.PageSize)).Limit(int(f.PageSize)).Find(&items).Error
	return items, total, err
}

// Prune deletes a bounded oldest batch; a zero retention policy never calls it.
func (r *Audit) Prune(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Exec("DELETE FROM audit_logs WHERE created_at < ? ORDER BY created_at, id LIMIT 1000", before).Error
}
