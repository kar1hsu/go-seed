package repository

import (
	"context"
	"time"

	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Auth struct{ db *gorm.DB }

func NewAuth(db *gorm.DB) *Auth { return &Auth{db: db} }
func (s *Auth) Transaction(ctx context.Context, fn func(*Auth) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return fn(NewAuth(tx)) })
}
func (s *Auth) ByUsername(ctx context.Context, name string) (*model.Administrator, error) {
	var a model.Administrator
	err := s.db.WithContext(ctx).Where("username = ?", name).First(&a).Error
	return &a, err
}
func (s *Auth) LockAdmin(ctx context.Context, id uint64) (*model.Administrator, error) {
	var a model.Administrator
	err := s.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&a, id).Error
	return &a, err
}
func (s *Auth) CreateAdmin(ctx context.Context, a *model.Administrator) error {
	return s.db.WithContext(ctx).Create(a).Error
}
func (s *Auth) CreateSession(ctx context.Context, a *model.AdminSession) error {
	return s.db.WithContext(ctx).Create(a).Error
}
func (s *Auth) Session(ctx context.Context, id string) (*model.AdminSession, error) {
	var a model.AdminSession
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&a).Error
	return &a, err
}
func (s *Auth) Token(ctx context.Context, digest, sid string) (*model.AdminRefreshToken, error) {
	var t model.AdminRefreshToken
	err := s.db.WithContext(ctx).Where("digest = ? AND session_id = ?", digest, sid).First(&t).Error
	return &t, err
}
func (s *Auth) CreateToken(ctx context.Context, t *model.AdminRefreshToken) error {
	return s.db.WithContext(ctx).Create(t).Error
}
func (s *Auth) ConsumeToken(ctx context.Context, digest string, now time.Time) error {
	return s.db.WithContext(ctx).Model(&model.AdminRefreshToken{}).Where("digest = ?", digest).Update("consumed_at", now).Error
}
func (s *Auth) Revoke(ctx context.Context, id string, now time.Time) error {
	return s.db.WithContext(ctx).Model(&model.AdminSession{}).Where("id = ? AND revoked_at IS NULL", id).Update("revoked_at", now).Error
}
func (s *Auth) Password(ctx context.Context, id uint64, hash string, now time.Time) error {
	if err := s.db.WithContext(ctx).Model(&model.Administrator{}).Where("id = ?", id).Updates(map[string]any{"password_hash": hash, "token_version": gorm.Expr("token_version + 1"), "updated_at": now}).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&model.AdminSession{}).Where("admin_id = ? AND revoked_at IS NULL", id).Update("revoked_at", now).Error
}
func (s *Auth) Current(ctx context.Context, id uint64, sid string, version uint64, now time.Time) (*model.Administrator, error) {
	var a model.Administrator
	err := s.db.WithContext(ctx).Table("administrators AS a").Select("a.*").Joins("JOIN admin_sessions AS s ON s.admin_id = a.id").Where("a.id = ? AND a.enabled = ? AND a.token_version = ? AND s.id = ? AND s.token_version = ? AND s.revoked_at IS NULL AND s.expires_at > ?", id, true, version, sid, version, now).Take(&a).Error
	return &a, err
}

func (s *Auth) Disable(ctx context.Context, id uint64, now time.Time) error {
	if err := s.db.WithContext(ctx).Model(&model.Administrator{}).Where("id = ?", id).Updates(map[string]any{
		"enabled":       false,
		"token_version": gorm.Expr("token_version + 1"),
		"updated_at":    now,
	}).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&model.AdminSession{}).Where("admin_id = ? AND revoked_at IS NULL", id).Update("revoked_at", now).Error
}
