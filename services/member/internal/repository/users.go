package repository

import (
	"context"
	"database/sql"
	"github.com/kar1hsu/go-seed/services/member/internal/model"
	"github.com/zeromicro/go-zero/core/trace"
	"gorm.io/gorm"
	"strings"
	"time"
)

func (s *Auth) Event(ctx context.Context, userID uint64, action, actorKind string, actorID uint64) error {
	return s.db.WithContext(ctx).Create(&model.AuditEvent{UserID: userID, Action: action, ActorKind: actorKind, ActorID: actorID, TraceID: trace.TraceIDFromContext(ctx), CreatedAt: time.Now().UTC()}).Error
}
func (s *Auth) TouchLogin(ctx context.Context, id uint64, now time.Time) error {
	// Login activity does not invalidate profile revisions.
	return s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).UpdateColumn("last_login_at", now).Error
}
func (s *Auth) Profile(ctx context.Context, id uint64, nickname string, now time.Time) error {
	return s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(map[string]any{"nickname": nickname, "updated_at": now}).Error
}
func (s *Auth) Status(ctx context.Context, id uint64, enabled bool, now time.Time) error {
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(map[string]any{"enabled": enabled, "token_version": gorm.Expr("token_version + 1"), "updated_at": now}).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&model.MemberSession{}).Where("user_id = ? AND revoked_at IS NULL", id).Update("revoked_at", now).Error
}
func (s *Auth) List(ctx context.Context, page, size int64, keyword, enabled string) ([]model.User, int64, error) {
	items := []model.User{}
	var total int64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx.Model(&model.User{})
		if keyword != "" {
			pattern := "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(keyword) + "%"
			q = q.Where("username LIKE ? ESCAPE '!' OR nickname LIKE ? ESCAPE '!'", pattern, pattern)
		}
		if enabled != "all" {
			q = q.Where("enabled = ?", enabled == "enabled")
		}
		if err := q.Count(&total).Error; err != nil {
			return err
		}
		return q.Select("id,username,nickname,enabled,created_at,updated_at,last_login_at").Order("id DESC").Offset(int((page - 1) * size)).Limit(int(size)).Find(&items).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	return items, total, err
}
func (s *Auth) Detail(ctx context.Context, id uint64) (*model.User, []model.AuditEvent, error) {
	var user model.User
	events := []model.AuditEvent{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("id,username,nickname,enabled,created_at,updated_at,last_login_at").First(&user, id).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", id).Order("id DESC").Limit(20).Find(&events).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	return &user, events, err
}
