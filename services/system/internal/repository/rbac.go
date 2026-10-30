package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RBAC struct{ db *gorm.DB }
type PolicySnapshot struct {
	Version     uint64
	Roles       []model.Role
	Permissions []model.Permission
	Memberships []model.AdminRole
	Grants      []model.RolePermission
}

func NewRBAC(db *gorm.DB) *RBAC { return &RBAC{db: db} }
func (r *RBAC) Auth() *Auth     { return NewAuth(r.db) }
func (r *RBAC) Transaction(ctx context.Context, write bool, fn func(*RBAC) error) error {
	return r.db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		tx := NewRBAC(db)
		if write {
			if _, err := tx.LockPolicy(ctx); err != nil {
				return err
			}
		}
		return fn(tx)
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: !write})
}
func (r *RBAC) LockPolicy(ctx context.Context) (*model.PolicyState, error) {
	var state model.PolicyState
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&state, 1).Error
	return &state, err
}
func (r *RBAC) Version(ctx context.Context) (uint64, error) {
	var state model.PolicyState
	err := r.db.WithContext(ctx).First(&state, 1).Error
	return state.Version, err
}
func (r *RBAC) Bump(ctx context.Context) error {
	return r.db.WithContext(ctx).Model(&model.PolicyState{}).Where("id = 1").Update("version", gorm.Expr("version + 1")).Error
}
func (r *RBAC) Snapshot(ctx context.Context) (*PolicySnapshot, error) {
	s := &PolicySnapshot{}
	v, err := r.Version(ctx)
	if err != nil {
		return nil, err
	}
	s.Version = v
	for _, dest := range []any{&s.Roles, &s.Permissions, &s.Memberships, &s.Grants} {
		if err = r.db.WithContext(ctx).Find(dest).Error; err != nil {
			return nil, err
		}
	}
	return s, nil
}
func filtered(db *gorm.DB, keyword, enabled string) *gorm.DB {
	if keyword != "" {
		db = db.Where("name LIKE ? ESCAPE '!' OR code LIKE ? ESCAPE '!'", literalLike(keyword), literalLike(keyword))
	}
	if enabled != "all" {
		db = db.Where("enabled = ?", enabled == "enabled")
	}
	return db
}
func (r *RBAC) ListAdmins(ctx context.Context, page, size int64, keyword, enabled string) ([]model.Administrator, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Administrator{})
	if keyword != "" {
		q = q.Where("username LIKE ? ESCAPE '!' OR display_name LIKE ? ESCAPE '!'", literalLike(keyword), literalLike(keyword))
	}
	if enabled != "all" {
		q = q.Where("enabled = ?", enabled == "enabled")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := []model.Administrator{}
	err := q.Order("id DESC").Offset(int((page - 1) * size)).Limit(int(size)).Find(&items).Error
	return items, total, err
}
func (r *RBAC) ListRoles(ctx context.Context, page, size int64, keyword, enabled string) ([]model.Role, int64, error) {
	q := filtered(r.db.WithContext(ctx).Model(&model.Role{}), keyword, enabled)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := []model.Role{}
	err := q.Order("built_in DESC, id DESC").Offset(int((page - 1) * size)).Limit(int(size)).Find(&items).Error
	return items, total, err
}
func (r *RBAC) Role(ctx context.Context, id uint64) (*model.Role, error) {
	var item model.Role
	err := r.db.WithContext(ctx).First(&item, id).Error
	return &item, err
}
func (r *RBAC) CreateRole(ctx context.Context, item *model.Role) error {
	return r.db.WithContext(ctx).Create(item).Error
}
func (r *RBAC) UpdateRole(ctx context.Context, item *model.Role) error {
	return r.db.WithContext(ctx).Model(&model.Role{}).Where("id = ?", item.ID).Updates(map[string]any{
		"name": item.Name, "description": item.Description, "enabled": item.Enabled, "updated_at": item.UpdatedAt,
	}).Error
}
func (r *RBAC) DeleteRole(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Role{}, id).Error
}
func (r *RBAC) UpdateAdmin(ctx context.Context, id uint64, name string, enabled bool, now time.Time) error {
	if !enabled {
		if err := r.Auth().Disable(ctx, id, now); err != nil {
			return err
		}
	}
	if err := r.db.WithContext(ctx).Model(&model.Administrator{}).Where("id = ?", id).Updates(map[string]any{"display_name": name, "enabled": enabled, "updated_at": now}).Error; err != nil {
		return err
	}
	return nil
}
func (r *RBAC) SetRoles(ctx context.Context, id uint64, roles []uint64, now time.Time) error {
	if err := r.db.WithContext(ctx).Where("admin_id = ?", id).Delete(&model.AdminRole{}).Error; err != nil {
		return err
	}
	for _, role := range roles {
		if err := r.db.WithContext(ctx).Create(&model.AdminRole{AdminID: id, RoleID: role}).Error; err != nil {
			return err
		}
	}
	return r.db.WithContext(ctx).Model(&model.Administrator{}).Where("id = ?", id).Update("updated_at", now).Error
}
func (r *RBAC) SetGrants(ctx context.Context, id uint64, permissions []string, now time.Time) error {
	if err := r.db.WithContext(ctx).Where("role_id = ?", id).Delete(&model.RolePermission{}).Error; err != nil {
		return err
	}
	for _, code := range permissions {
		if err := r.db.WithContext(ctx).Create(&model.RolePermission{RoleID: id, PermissionCode: code}).Error; err != nil {
			return err
		}
	}
	return r.db.WithContext(ctx).Model(&model.Role{}).Where("id = ?", id).Update("updated_at", now).Error
}
func (r *RBAC) EnabledSuperCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("administrators a").Joins("JOIN admin_roles ar ON ar.admin_id=a.id JOIN roles r ON r.id=ar.role_id").Where("a.enabled = ? AND r.built_in = ? AND r.enabled = ?", true, true, true).Count(&count).Error
	return count, err
}
func (r *RBAC) SyncPermissions(ctx context.Context, items []model.Permission) error {
	if err := r.db.WithContext(ctx).Model(&model.Permission{}).Where("active = ?", true).Update("active", false).Error; err != nil {
		return err
	}
	for _, item := range items {
		if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "group_name", "kind", "active"})}).Create(&item).Error; err != nil {
			return err
		}
	}
	return r.Bump(ctx)
}
func (r *RBAC) MarkBootstrapped(ctx context.Context) error {
	return r.db.WithContext(ctx).Model(&model.PolicyState{}).Where("id = 1").Update("bootstrapped", true).Error
}

func (r *RBAC) Admin(ctx context.Context, id uint64) (*model.Administrator, error) {
	var a model.Administrator
	err := r.db.WithContext(ctx).First(&a, id).Error
	return &a, err
}

func literalLike(s string) string {
	return "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(s) + "%"
}
