package management

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/internal/application/auth"
	"github.com/kar1hsu/go-seed/services/system/internal/audit"
	"github.com/kar1hsu/go-seed/services/system/internal/authorization"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/resources"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Service struct {
	store      *repository.RBAC
	authorizer *authorization.Authorizer
	catalog    []resources.Permission
}
type Query struct {
	Page, PageSize   int64
	Keyword, Enabled string
}
type Administrator struct {
	model.Administrator
	Roles      []model.Role
	Manageable bool
}
type Role struct {
	model.Role
	Permissions            []string
	MemberCount            int64
	Manageable, Assignable bool
}

func New(store *repository.RBAC, authorizer *authorization.Authorizer) (*Service, error) {
	catalog, err := resources.Catalog()
	if err != nil {
		return nil, err
	}
	return &Service{store, authorizer, catalog}, nil
}
func publicError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return xerr.New(codes.NotFound)
	}
	var duplicate *mysql.MySQLError
	if errors.As(err, &duplicate) && duplicate.Number == 1062 {
		return xerr.New(codes.AlreadyExists)
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	logx.WithContext(ctx).Error(fmt.Sprintf("management storage failure (%T)", err))
	return xerr.New(codes.Unavailable)
}
func (s *Service) run(ctx context.Context, raw, permission string, write bool, fn func(*repository.RBAC, *authorization.Access) error) error {
	return s.runTransaction(ctx, raw, permission, write, write, fn)
}

// Settings writes share the policy lock with revocations but do not change the authorization version.
func (s *Service) runData(ctx context.Context, raw, permission string, write bool, fn func(*repository.RBAC, *authorization.Access) error) error {
	return s.runTransaction(ctx, raw, permission, write, false, fn)
}
func (s *Service) runTransaction(ctx context.Context, raw, permission string, write, policyChanged bool, fn func(*repository.RBAC, *authorization.Access) error) error {
	return publicError(ctx, s.store.Transaction(ctx, write, func(tx *repository.RBAC) error {
		access, err := s.authorizer.Current(ctx, tx, raw)
		if err != nil {
			return err
		}
		if permission != "" {
			if err = access.Require(permission); err != nil {
				return err
			}
		}
		if err = fn(tx, access); err != nil {
			return err
		}
		if policyChanged {
			return tx.Bump(ctx)
		}
		return nil
	}))
}
func (s *Service) Current(ctx context.Context, raw string) (*authorization.Access, error) {
	var result *authorization.Access
	err := s.run(ctx, raw, "", false, func(_ *repository.RBAC, a *authorization.Access) error { result = a; return nil })
	return result, err
}
func validateQuery(q Query) error {
	if q.Page < 1 || q.Page > 100000 || q.PageSize < 1 || q.PageSize > 100 || len([]rune(q.Keyword)) > 64 || (q.Enabled != "all" && q.Enabled != "enabled" && q.Enabled != "disabled") {
		return xerr.New(codes.InvalidArgument)
	}
	return nil
}
func ID(raw string) (uint64, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 || strconv.FormatUint(id, 10) != raw {
		return 0, xerr.New(codes.InvalidArgument)
	}
	return id, nil
}
func Revision(t time.Time) string { return t.UTC().Format(time.RFC3339Nano) }
func revisionMatches(raw string, t time.Time) error {
	if raw != Revision(t) {
		return xerr.New(codes.Aborted)
	}
	return nil
}
func validName(value string) bool { return strings.TrimSpace(value) != "" && len([]rune(value)) <= 64 }
func nextTime(previous time.Time) time.Time {
	now := time.Now().UTC().Truncate(time.Microsecond)
	if !now.After(previous) {
		return previous.Add(time.Microsecond)
	}
	return now
}
func (s *Service) ListAdministrators(ctx context.Context, raw string, q Query) ([]Administrator, int64, error) {
	if err := validateQuery(q); err != nil {
		return nil, 0, err
	}
	result := []Administrator{}
	var total int64
	err := s.run(ctx, raw, resources.UserRead, false, func(tx *repository.RBAC, a *authorization.Access) error {
		items, count, err := tx.ListAdmins(ctx, q.Page, q.PageSize, q.Keyword, q.Enabled)
		if err != nil {
			return err
		}
		total = count
		for _, item := range items {
			result = append(result, Administrator{item, a.Policy.RolesFor(item.ID), a.CanManageAdmin(item.ID)})
		}
		return nil
	})
	return result, total, err
}
func (s *Service) CreateAdministrator(ctx context.Context, raw, username, name, password string) (*Administrator, error) {
	if !auth.ValidUsername(username) || !validName(name) || !auth.ValidPassword(password) {
		return nil, xerr.New(codes.InvalidArgument)
	}
	// Reject unauthorized callers before the expensive password hash, then recheck inside the write transaction.
	if err := s.run(ctx, raw, resources.UserCreate, false, func(*repository.RBAC, *authorization.Access) error { return nil }); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, publicError(ctx, err)
	}
	now := nextTime(time.Time{})
	item := model.Administrator{Username: username, DisplayName: strings.TrimSpace(name), PasswordHash: string(hash), Enabled: true, TokenVersion: 1, CreatedAt: now, UpdatedAt: now}
	err = s.run(ctx, raw, resources.UserCreate, true, func(tx *repository.RBAC, a *authorization.Access) error {
		if err := tx.Auth().CreateAdmin(ctx, &item); err != nil {
			return err
		}
		return appendAudit(ctx, tx, a, resources.UserCreate, "administrator", item.ID, item.Username, item.DisplayName, nil, audit.Administrator(item))
	})
	return &Administrator{Administrator: item, Roles: []model.Role{}, Manageable: true}, err
}
func (s *Service) UpdateAdministrator(ctx context.Context, raw string, id uint64, name string, enabled bool, revision string) error {
	if id == 0 || !validName(name) {
		return xerr.New(codes.InvalidArgument)
	}
	return s.run(ctx, raw, resources.UserUpdate, true, func(tx *repository.RBAC, a *authorization.Access) error {
		if !a.CanManageAdmin(id) {
			return xerr.New(codes.PermissionDenied)
		}
		target, err := tx.Auth().LockAdmin(ctx, id)
		if err != nil {
			return err
		}
		if err = revisionMatches(revision, target.UpdatedAt); err != nil {
			return err
		}
		if !enabled && id == a.Admin.ID {
			return xerr.New(codes.FailedPrecondition)
		}
		if err = tx.UpdateAdmin(ctx, id, strings.TrimSpace(name), enabled, nextTime(target.UpdatedAt)); err != nil {
			return err
		}
		before := audit.Administrator(*target)
		target.DisplayName, target.Enabled = strings.TrimSpace(name), enabled
		return appendAudit(ctx, tx, a, resources.UserUpdate, "administrator", id, target.Username, target.DisplayName, before, audit.Administrator(*target))
	})
}
func (s *Service) AssignRoles(ctx context.Context, raw string, id uint64, ids []uint64, revision string) error {
	if id == 0 || len(ids) > 100 {
		return xerr.New(codes.InvalidArgument)
	}
	return s.run(ctx, raw, resources.UserAssignRole, true, func(tx *repository.RBAC, a *authorization.Access) error {
		if !a.CanManageAdmin(id) || id == a.Admin.ID && !a.SuperAdmin {
			return xerr.New(codes.PermissionDenied)
		}
		target, err := tx.Auth().LockAdmin(ctx, id)
		if err != nil {
			return err
		}
		if err = revisionMatches(revision, target.UpdatedAt); err != nil {
			return err
		}
		seen := map[uint64]bool{}
		roles := []model.Role{}
		hasSuper := false
		for _, roleID := range ids {
			if roleID == 0 || seen[roleID] {
				return xerr.New(codes.InvalidArgument)
			}
			seen[roleID] = true
			role, err := tx.Role(ctx, roleID)
			if err != nil {
				return err
			}
			if !a.CanAssignRole(*role) {
				return xerr.New(codes.PermissionDenied)
			}
			hasSuper = hasSuper || role.BuiltIn
			roles = append(roles, *role)
		}
		wasSuper := false
		for _, role := range a.Policy.RolesFor(id) {
			wasSuper = wasSuper || role.BuiltIn && role.Enabled
		}
		if target.Enabled && wasSuper && !hasSuper {
			count, err := tx.EnabledSuperCount(ctx)
			if err != nil {
				return err
			}
			if count <= 1 {
				return xerr.New(codes.FailedPrecondition)
			}
		}
		if err = tx.SetRoles(ctx, id, ids, nextTime(target.UpdatedAt)); err != nil {
			return err
		}
		return appendAudit(ctx, tx, a, resources.UserAssignRole, "administrator", id, target.Username, target.DisplayName, audit.Roles(a.Policy.RolesFor(id)), audit.Roles(roles))
	})
}
func roleView(role model.Role, a *authorization.Access) Role {
	result := Role{Role: role, Permissions: a.Policy.RolePermissions(role.ID), Manageable: a.CanManageRole(role), Assignable: a.CanAssignRole(role)}
	for _, member := range a.Policy.Snapshot.Memberships {
		if member.RoleID == role.ID {
			result.MemberCount++
		}
	}
	return result
}
func (s *Service) ListRoles(ctx context.Context, raw string, q Query) ([]Role, int64, error) {
	if err := validateQuery(q); err != nil {
		return nil, 0, err
	}
	result := []Role{}
	var total int64
	err := s.run(ctx, raw, resources.RoleRead, false, func(tx *repository.RBAC, a *authorization.Access) error {
		items, count, err := tx.ListRoles(ctx, q.Page, q.PageSize, q.Keyword, q.Enabled)
		if err != nil {
			return err
		}
		total = count
		for _, item := range items {
			result = append(result, roleView(item, a))
		}
		return nil
	})
	return result, total, err
}
func (s *Service) CreateRole(ctx context.Context, raw, code, name, description string) (*Role, error) {
	if !regexp.MustCompile(`^[a-z][a-z0-9_]{2,63}$`).MatchString(code) || code == resources.SuperRole || !validName(name) || len([]rune(description)) > 255 {
		return nil, xerr.New(codes.InvalidArgument)
	}
	now := nextTime(time.Time{})
	item := model.Role{Code: code, Name: strings.TrimSpace(name), Description: description, Enabled: true, CreatedAt: now, UpdatedAt: now}
	err := s.run(ctx, raw, resources.RoleCreate, true, func(tx *repository.RBAC, a *authorization.Access) error {
		if err := tx.CreateRole(ctx, &item); err != nil {
			return err
		}
		return appendAudit(ctx, tx, a, resources.RoleCreate, "role", item.ID, item.Code, item.Name, nil, audit.Role(item))
	})
	return &Role{Role: item, Permissions: []string{}, Manageable: true, Assignable: true}, err
}
func (s *Service) UpdateRole(ctx context.Context, raw string, id uint64, name, description string, enabled bool, revision string) error {
	if id == 0 || !validName(name) || len([]rune(description)) > 255 {
		return xerr.New(codes.InvalidArgument)
	}
	return s.run(ctx, raw, resources.RoleUpdate, true, func(tx *repository.RBAC, a *authorization.Access) error {
		role, err := tx.Role(ctx, id)
		if err != nil {
			return err
		}
		if !a.CanManageRole(*role) {
			return xerr.New(codes.PermissionDenied)
		}
		if err = revisionMatches(revision, role.UpdatedAt); err != nil {
			return err
		}
		before := audit.Role(*role)
		role.Name, role.Description, role.Enabled, role.UpdatedAt = strings.TrimSpace(name), description, enabled, nextTime(role.UpdatedAt)
		if err = tx.UpdateRole(ctx, role); err != nil {
			return err
		}
		return appendAudit(ctx, tx, a, resources.RoleUpdate, "role", id, role.Code, role.Name, before, audit.Role(*role))
	})
}
func (s *Service) GrantRole(ctx context.Context, raw string, id uint64, permissions []string, revision string) error {
	if id == 0 || len(permissions) > 1000 {
		return xerr.New(codes.InvalidArgument)
	}
	return s.run(ctx, raw, resources.RoleGrant, true, func(tx *repository.RBAC, a *authorization.Access) error {
		role, err := tx.Role(ctx, id)
		if err != nil {
			return err
		}
		if !a.CanManageRole(*role) || !a.Covers(permissions) {
			return xerr.New(codes.PermissionDenied)
		}
		if err = revisionMatches(revision, role.UpdatedAt); err != nil {
			return err
		}
		selected := map[string]bool{}
		known := map[string]resources.Permission{}
		for _, p := range s.catalog {
			known[p.Code] = p
		}
		for _, code := range permissions {
			if selected[code] {
				return xerr.New(codes.InvalidArgument)
			}
			if _, ok := known[code]; !ok {
				return xerr.New(codes.InvalidArgument)
			}
			selected[code] = true
		}
		for _, code := range permissions {
			for _, dependency := range known[code].Requires {
				if !selected[dependency] {
					return xerr.New(codes.InvalidArgument)
				}
			}
		}
		if err = tx.SetGrants(ctx, id, permissions, nextTime(role.UpdatedAt)); err != nil {
			return err
		}
		return appendAudit(ctx, tx, a, resources.RoleGrant, "role", id, role.Code, role.Name, audit.Permissions(assignedPermissions(a.Policy, id)), audit.Permissions(permissions))
	})
}
func (s *Service) DeleteRole(ctx context.Context, raw string, id uint64, revision string) error {
	if id == 0 {
		return xerr.New(codes.InvalidArgument)
	}
	return s.run(ctx, raw, resources.RoleDelete, true, func(tx *repository.RBAC, a *authorization.Access) error {
		role, err := tx.Role(ctx, id)
		if err != nil {
			return err
		}
		if !a.CanManageRole(*role) {
			return xerr.New(codes.PermissionDenied)
		}
		if err = revisionMatches(revision, role.UpdatedAt); err != nil {
			return err
		}
		for _, member := range a.Policy.Snapshot.Memberships {
			if member.RoleID == id {
				return xerr.New(codes.FailedPrecondition)
			}
		}
		if err = tx.DeleteRole(ctx, id); err != nil {
			return err
		}
		before := append(audit.Role(*role), audit.Permissions(assignedPermissions(a.Policy, id))...)
		return appendAudit(ctx, tx, a, resources.RoleDelete, "role", id, role.Code, role.Name, before, nil)
	})
}
func (s *Service) Permissions(ctx context.Context, raw string) ([]resources.Permission, error) {
	result := []resources.Permission{}
	err := s.run(ctx, raw, resources.RoleRead, false, func(_ *repository.RBAC, a *authorization.Access) error {
		for _, entry := range s.catalog {
			for _, p := range a.Policy.Snapshot.Permissions {
				if p.Active && p.Code == entry.Code {
					result = append(result, entry)
				}
			}
		}
		return nil
	})
	return result, err
}
func (s *Service) SyncResources(ctx context.Context) error {
	return publicError(ctx, s.store.Transaction(ctx, true, func(tx *repository.RBAC) error {
		items := []model.Permission{}
		for _, p := range s.catalog {
			items = append(items, model.Permission{Code: p.Code, Name: p.Name, GroupName: p.Group, Kind: p.Kind, Active: true})
		}
		return tx.SyncPermissions(ctx, items)
	}))
}

// BootstrapRoot grants the built-in role only once, never restoring removed memberships on later runs.
func (s *Service) BootstrapRoot(ctx context.Context, username string) (bool, error) {
	created := false
	err := s.store.Transaction(ctx, true, func(tx *repository.RBAC) error {
		state, err := tx.LockPolicy(ctx)
		if err != nil {
			return err
		}
		if state.Bootstrapped {
			return nil
		}
		admin, err := tx.Auth().ByUsername(ctx, username)
		if err != nil {
			return err
		}
		if !admin.Enabled {
			return xerr.New(codes.FailedPrecondition)
		}
		now := nextTime(time.Time{})
		role := model.Role{Code: resources.SuperRole, Name: "超级管理员", Description: "内置系统管理角色", Enabled: true, BuiltIn: true, CreatedAt: now, UpdatedAt: now}
		if err = tx.CreateRole(ctx, &role); err != nil {
			return err
		}
		if err = tx.SetRoles(ctx, admin.ID, []uint64{role.ID}, nextTime(admin.UpdatedAt)); err != nil {
			return err
		}
		if err = tx.MarkBootstrapped(ctx); err != nil {
			return err
		}
		created = true
		return tx.Bump(ctx)
	})
	return created, publicError(ctx, err)
}
