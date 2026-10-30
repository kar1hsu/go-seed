package authorization

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/casbin/casbin/v3"
	casbinmodel "github.com/casbin/casbin/v3/model"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/resources"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

type Policy struct {
	Snapshot *repository.PolicySnapshot
	enforcer *casbin.SyncedEnforcer
}
type Access struct {
	Admin       *model.Administrator
	Permissions []string
	Roles       []string
	SuperAdmin  bool
	Policy      *Policy
}
type Authorizer struct {
	tokens *identity.Tokens
	policy atomic.Pointer[Policy]
	mu     sync.Mutex
}

func New(tokens *identity.Tokens) *Authorizer { return &Authorizer{tokens: tokens} }
func (a *Authorizer) Current(ctx context.Context, tx *repository.RBAC, raw string) (*Access, error) {
	claims, err := a.tokens.Parse(raw)
	if err != nil {
		return nil, xerr.New(codes.Unauthenticated)
	}
	id, _ := strconv.ParseUint(claims.Subject, 10, 64)
	admin, err := tx.Auth().Current(ctx, id, claims.SessionID, claims.Version, time.Now().UTC())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, xerr.New(codes.Unauthenticated)
	}
	if err != nil {
		return nil, err
	}
	return a.access(ctx, tx, admin)
}

// Deferred rechecks current account status and policy without storing or extending a login token.
func (a *Authorizer) Deferred(ctx context.Context, tx *repository.RBAC, id uint64) (*Access, error) {
	admin, err := tx.Admin(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, xerr.New(codes.PermissionDenied)
	}
	if err != nil {
		return nil, err
	}
	if !admin.Enabled {
		return nil, xerr.New(codes.PermissionDenied)
	}
	return a.access(ctx, tx, admin)
}
func (a *Authorizer) access(ctx context.Context, tx *repository.RBAC, admin *model.Administrator) (*Access, error) {
	version, err := tx.Version(ctx)
	if err != nil {
		return nil, err
	}
	policy := a.policy.Load()
	if policy == nil || policy.Snapshot.Version != version {
		policy, err = a.load(ctx, tx, version)
		if err != nil {
			return nil, err
		}
	}
	access := &Access{Admin: admin, Permissions: []string{}, Roles: []string{}, Policy: policy}
	for _, role := range policy.RolesFor(admin.ID) {
		if role.Enabled {
			access.Roles = append(access.Roles, role.Code)
			access.SuperAdmin = access.SuperAdmin || role.BuiltIn && role.Code == resources.SuperRole
		}
	}
	for _, permission := range policy.Snapshot.Permissions {
		if !permission.Active {
			continue
		}
		ok, err := policy.enforcer.Enforce(userKey(admin.ID), permission.Code)
		if err != nil {
			return nil, err
		}
		if ok {
			access.Permissions = append(access.Permissions, permission.Code)
		}
	}
	sort.Strings(access.Roles)
	sort.Strings(access.Permissions)
	return access, nil
}
func (a *Authorizer) load(ctx context.Context, tx *repository.RBAC, version uint64) (*Policy, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if current := a.policy.Load(); current != nil && current.Snapshot.Version == version {
		return current, nil
	}
	snapshot, err := tx.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	m, err := casbinmodel.NewModelFromString(resources.CasbinModel)
	if err != nil {
		return nil, err
	}
	e, err := casbin.NewSyncedEnforcer(m)
	if err != nil {
		return nil, err
	}
	e.EnableAutoSave(false)
	p := &Policy{Snapshot: snapshot, enforcer: e}
	for _, role := range snapshot.Roles {
		if !role.Enabled {
			continue
		}
		for _, permission := range p.RolePermissions(role.ID) {
			if _, err = e.AddPolicy(roleKey(role.ID), permission); err != nil {
				return nil, err
			}
		}
	}
	for _, member := range snapshot.Memberships {
		if _, err = e.AddGroupingPolicy(userKey(member.AdminID), roleKey(member.RoleID)); err != nil {
			return nil, err
		}
	}
	if current := a.policy.Load(); current == nil || current.Snapshot.Version <= snapshot.Version {
		a.policy.Store(p)
	}
	return p, nil
}
func userKey(id uint64) string { return "user:" + strconv.FormatUint(id, 10) }
func roleKey(id uint64) string { return "role:" + strconv.FormatUint(id, 10) }
func (a *Access) Has(permission string) bool {
	for _, code := range a.Permissions {
		if code == permission {
			return true
		}
	}
	return false
}
func (a *Access) Require(permission string) error {
	if !a.Has(permission) {
		return xerr.New(codes.PermissionDenied)
	}
	return nil
}
func (a *Access) Covers(permissions []string) bool {
	for _, code := range permissions {
		if !a.Has(code) {
			return false
		}
	}
	return true
}
func (p *Policy) RolesFor(id uint64) []model.Role {
	result := []model.Role{}
	for _, member := range p.Snapshot.Memberships {
		if member.AdminID == id {
			for _, role := range p.Snapshot.Roles {
				if role.ID == member.RoleID {
					result = append(result, role)
				}
			}
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
func (p *Policy) RolePermissions(id uint64) []string {
	result := []string{}
	super := false
	for _, role := range p.Snapshot.Roles {
		if role.ID == id {
			super = role.BuiltIn && role.Code == resources.SuperRole
		}
	}
	for _, permission := range p.Snapshot.Permissions {
		if !permission.Active {
			continue
		}
		if super {
			result = append(result, permission.Code)
			continue
		}
		for _, grant := range p.Snapshot.Grants {
			if grant.RoleID == id && grant.PermissionCode == permission.Code {
				result = append(result, permission.Code)
			}
		}
	}
	sort.Strings(result)
	return result
}
func (a *Access) CanManageRole(role model.Role) bool {
	return !role.BuiltIn && (a.SuperAdmin || a.Covers(a.Policy.RolePermissions(role.ID)))
}
func (a *Access) CanAssignRole(role model.Role) bool {
	return role.Enabled && (a.SuperAdmin || !role.BuiltIn && a.Covers(a.Policy.RolePermissions(role.ID)))
}
func (a *Access) CanManageAdmin(id uint64) bool {
	if a.SuperAdmin {
		return true
	}
	for _, role := range a.Policy.RolesFor(id) {
		if role.BuiltIn || !a.Covers(a.Policy.RolePermissions(role.ID)) {
			return false
		}
	}
	return true
}
