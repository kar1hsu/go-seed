//go:build integration

package management

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/auth"
	"github.com/kar1hsu/go-seed/services/system/internal/authorization"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/migrations"
	"github.com/kar1hsu/go-seed/services/system/resources"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

const password = "rbac-integration-password"

var ctx = context.Background()

type fixture struct {
	db              *gorm.DB
	a, b            *Service
	auth            *auth.Service
	root            string
	rootID, superID uint64
}

func require(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func code(t *testing.T, err error, want codes.Code) {
	t.Helper()
	if status.Code(err) != want {
		t.Fatalf("got %v, want %v", err, want)
	}
}
func setup(t *testing.T) *fixture {
	t.Helper()
	port, err := strconv.Atoi(os.Getenv("MYSQL_PORT"))
	require(t, err)
	config := dbx.MySQLConfig{Host: os.Getenv("MYSQL_HOST"), Port: port, Database: "mysql", User: "root", Password: os.Getenv("MYSQL_ROOT_PASSWORD"), MaxOpenConns: 12, MaxIdleConns: 2, ConnMaxLifetime: time.Minute}
	admin, err := dbx.OpenMySQL(config)
	require(t, err)
	adminPool, _ := admin.DB()
	t.Cleanup(func() { adminPool.Close() })
	suffix := make([]byte, 10)
	_, err = rand.Read(suffix)
	require(t, err)
	name := "goseed_rbac_test_" + hex.EncodeToString(suffix)
	require(t, admin.Exec("CREATE DATABASE `"+name+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error)
	t.Cleanup(func() {
		if err := admin.Exec("DROP DATABASE `" + name + "`").Error; err != nil {
			t.Error(err)
		}
	})
	config.Database = name
	db, err := dbx.OpenMySQL(config)
	require(t, err)
	pool, _ := db.DB()
	t.Cleanup(func() { pool.Close() })
	require(t, migrations.Up(ctx, pool))
	require(t, migrations.Up(ctx, pool))
	tokens, err := identity.New(identity.Config{Secret: strings.Repeat("r", 48), Issuer: "test-rbac", Audience: "test-api", AccessTTL: time.Hour})
	require(t, err)
	as, err := auth.New(repository.NewAuth(db), tokens, 24*time.Hour, func(context.Context, string) error { return nil })
	require(t, err)
	makeService := func() *Service {
		s, err := New(repository.NewRBAC(db), authorization.New(tokens), nil)
		require(t, err)
		return s
	}
	f := &fixture{db: db, a: makeService(), b: makeService(), auth: as}
	_, err = as.Bootstrap(ctx, "test_root", password, "Test root")
	require(t, err)
	require(t, f.a.SyncResources(ctx))
	created, err := f.a.BootstrapRoot(ctx, "test_root")
	require(t, err)
	if !created {
		t.Fatal("root not bootstrapped")
	}
	session, err := as.Login(ctx, "test_root", password)
	require(t, err)
	f.root = session.AccessToken
	root, err := repository.NewAuth(db).ByUsername(ctx, "test_root")
	require(t, err)
	f.rootID = root.ID
	var super model.Role
	require(t, db.Where("code = ?", resources.SuperRole).First(&super).Error)
	f.superID = super.ID
	return f
}
func (f *fixture) user(t *testing.T, name string) (uint64, string) {
	t.Helper()
	a, err := f.a.CreateAdministrator(ctx, f.root, name, name, password)
	require(t, err)
	s, err := f.auth.Login(ctx, name, password)
	require(t, err)
	return a.ID, s.AccessToken
}
func (f *fixture) role(t *testing.T, name string, permissions ...string) uint64 {
	t.Helper()
	r, err := f.a.CreateRole(ctx, f.root, name, name, "")
	require(t, err)
	require(t, f.a.GrantRole(ctx, f.root, r.ID, permissions, Revision(r.UpdatedAt)))
	return r.ID
}
func (f *fixture) adminRevision(t *testing.T, id uint64) string {
	t.Helper()
	var a model.Administrator
	require(t, f.db.First(&a, id).Error)
	return Revision(a.UpdatedAt)
}
func (f *fixture) roleRevision(t *testing.T, id uint64) string {
	t.Helper()
	var a model.Role
	require(t, f.db.First(&a, id).Error)
	return Revision(a.UpdatedAt)
}
func (f *fixture) assign(t *testing.T, id uint64, roles ...uint64) {
	t.Helper()
	require(t, f.a.AssignRoles(ctx, f.root, id, roles, f.adminRevision(t, id)))
}
func TestManagementAuthorizationAndRevocation(t *testing.T) {
	f := setup(t)
	readerID, reader := f.user(t, "reader")
	readRole := f.role(t, "reader_role", resources.UserRead)
	f.assign(t, readerID, readRole)
	query := Query{Page: 1, PageSize: 1, Enabled: "all"}
	rows, total, err := f.b.ListAdministrators(ctx, reader, query)
	require(t, err)
	if total != 2 || len(rows) != 1 {
		t.Fatalf("pagination %d/%d", len(rows), total)
	}
	_, _, err = f.b.ListRoles(ctx, reader, query)
	code(t, err, codes.PermissionDenied)
	_, err = f.b.CreateRole(ctx, reader, "denied", "Denied", "")
	code(t, err, codes.PermissionDenied)
	code(t, f.b.UpdateAdministrator(ctx, reader, f.rootID, "forged", true, f.adminRevision(t, f.rootID)), codes.PermissionDenied)
	query.Page = 0
	_, _, err = f.a.ListAdministrators(ctx, f.root, query)
	code(t, err, codes.InvalidArgument)
	// Same JWT, different process: committed policy revocation is immediately observed.
	require(t, f.a.GrantRole(ctx, f.root, readRole, []string{}, f.roleRevision(t, readRole)))
	_, _, err = f.b.ListAdministrators(ctx, reader, Query{Page: 1, PageSize: 20, Enabled: "all"})
	code(t, err, codes.PermissionDenied)
	require(t, f.a.GrantRole(ctx, f.root, readRole, []string{resources.UserRead}, f.roleRevision(t, readRole)))
	current, err := f.b.Current(ctx, reader)
	require(t, err)
	if !current.Has(resources.UserRead) {
		t.Fatal("grant not visible")
	}
	// Disabled roles remove access without changing the session or stored grants.
	require(t, f.a.UpdateRole(ctx, f.root, readRole, "reader_role", "", false, f.roleRevision(t, readRole)))
	current, err = f.b.Current(ctx, reader)
	require(t, err)
	if len(current.Permissions) != 0 {
		t.Fatal("disabled role allowed")
	}
	require(t, f.a.UpdateRole(ctx, f.root, readRole, "reader_role", "", true, f.roleRevision(t, readRole)))
	old := f.adminRevision(t, readerID)
	require(t, f.a.UpdateAdministrator(ctx, f.root, readerID, "Renamed", true, old))
	code(t, f.a.UpdateAdministrator(ctx, f.root, readerID, "Stale", true, old), codes.Aborted)
	require(t, f.a.UpdateAdministrator(ctx, f.root, readerID, "Renamed", false, f.adminRevision(t, readerID)))
	_, err = f.b.Current(ctx, reader)
	code(t, err, codes.Unauthenticated)
	require(t, f.a.UpdateAdministrator(ctx, f.root, readerID, "Renamed", true, f.adminRevision(t, readerID)))
	_, err = f.b.Current(ctx, reader)
	code(t, err, codes.Unauthenticated)
}
func TestDelegationScopeAndRoleLifecycle(t *testing.T) {
	f := setup(t)
	operatorID, operator := f.user(t, "operator")
	operatorRole := f.role(t, "operator_role", resources.UserRead, resources.UserUpdate, resources.UserAssignRole, resources.RoleRead, resources.RoleGrant, resources.RoleUpdate, resources.RoleDelete)
	f.assign(t, operatorID, operatorRole)
	lowID, _ := f.user(t, "low_user")
	lowRole := f.role(t, "low_role", resources.UserRead)
	require(t, f.b.AssignRoles(ctx, operator, lowID, []uint64{lowRole}, f.adminRevision(t, lowID)))
	code(t, f.b.AssignRoles(ctx, operator, operatorID, []uint64{operatorRole}, f.adminRevision(t, operatorID)), codes.PermissionDenied)
	code(t, f.b.AssignRoles(ctx, operator, lowID, []uint64{f.superID}, f.adminRevision(t, lowID)), codes.PermissionDenied)
	code(t, f.b.UpdateAdministrator(ctx, operator, f.rootID, "Attack", true, f.adminRevision(t, f.rootID)), codes.PermissionDenied)
	code(t, f.b.GrantRole(ctx, operator, lowRole, []string{resources.RoleRead, resources.RoleCreate}, f.roleRevision(t, lowRole)), codes.PermissionDenied)
	code(t, f.a.GrantRole(ctx, f.root, lowRole, []string{resources.UserCreate}, f.roleRevision(t, lowRole)), codes.InvalidArgument)
	code(t, f.a.GrantRole(ctx, f.root, lowRole, []string{resources.UserRead, resources.UserRead}, f.roleRevision(t, lowRole)), codes.InvalidArgument)
	code(t, f.a.DeleteRole(ctx, f.root, lowRole, f.roleRevision(t, lowRole)), codes.FailedPrecondition)
	f.assign(t, lowID)
	revision := f.roleRevision(t, lowRole)
	require(t, f.b.GrantRole(ctx, operator, lowRole, []string{resources.RoleRead}, revision))
	code(t, f.a.DeleteRole(ctx, f.root, lowRole, revision), codes.Aborted)
	require(t, f.b.DeleteRole(ctx, operator, lowRole, f.roleRevision(t, lowRole)))
	_, err := f.a.CreateRole(ctx, f.root, "operator_role", "Duplicate", "")
	code(t, err, codes.AlreadyExists)
	// A disabled role still counts when comparing authority, so re-enabling cannot escalate.
	highRole := f.role(t, "high_role", resources.UserRead, resources.UserCreate)
	require(t, f.a.UpdateRole(ctx, f.root, highRole, "high_role", "", false, f.roleRevision(t, highRole)))
	code(t, f.b.UpdateRole(ctx, operator, highRole, "high_role", "", true, f.roleRevision(t, highRole)), codes.PermissionDenied)
	code(t, f.a.AssignRoles(ctx, f.root, lowID, []uint64{highRole}, f.adminRevision(t, lowID)), codes.PermissionDenied)
}
func TestLastSuperAdministratorAndBootstrap(t *testing.T) {
	f := setup(t)
	for _, err := range []error{
		f.a.GrantRole(ctx, f.root, f.superID, []string{}, f.roleRevision(t, f.superID)),
		f.a.DeleteRole(ctx, f.root, f.superID, f.roleRevision(t, f.superID)),
		f.a.UpdateRole(ctx, f.root, f.superID, "Changed", "", false, f.roleRevision(t, f.superID)),
	} {
		code(t, err, codes.PermissionDenied)
	}
	code(t, f.a.AssignRoles(ctx, f.root, f.rootID, nil, f.adminRevision(t, f.rootID)), codes.FailedPrecondition)
	code(t, f.auth.Disable(ctx, "test_root"), codes.FailedPrecondition)
	secondID, second := f.user(t, "second_root")
	f.assign(t, secondID, f.superID)
	firstRevision, secondRevision := f.adminRevision(t, f.rootID), f.adminRevision(t, secondID)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, call := range []func() error{
		func() error { return f.a.AssignRoles(ctx, f.root, f.rootID, nil, firstRevision) },
		func() error { return f.b.AssignRoles(ctx, second, secondID, nil, secondRevision) },
	} {
		wg.Add(1)
		go func(fn func() error) { defer wg.Done(); results <- fn() }(call)
	}
	wg.Wait()
	close(results)
	success, denied := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if status.Code(err) == codes.FailedPrecondition {
			denied++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || denied != 1 {
		t.Fatalf("last root race successes=%d denied=%d", success, denied)
	}
	// The first bootstrap account can lose its role; repeating bootstrap must not silently restore it.
	a, err := f.a.Current(ctx, f.root)
	require(t, err)
	if a.SuperAdmin {
		require(t, f.a.AssignRoles(ctx, f.root, secondID, []uint64{f.superID}, f.adminRevision(t, secondID)))
		require(t, f.a.AssignRoles(ctx, f.root, f.rootID, nil, f.adminRevision(t, f.rootID)))
	}
	created, err := f.b.BootstrapRoot(ctx, "test_root")
	require(t, err)
	if created {
		t.Fatal("bootstrap restored removed role")
	}
	a, err = f.b.Current(ctx, f.root)
	require(t, err)
	if a.SuperAdmin {
		t.Fatal("bootstrap escalated account")
	}
	code(t, f.auth.Disable(ctx, "second_root"), codes.FailedPrecondition)
}
func TestResourceSyncPreservesGrantsAndFailsClosed(t *testing.T) {
	f := setup(t)
	userID, user := f.user(t, "sync_user")
	roleID := f.role(t, "sync_role", resources.UserRead)
	f.assign(t, userID, roleID)
	before, err := f.b.Current(ctx, user)
	require(t, err)
	f.a.catalog = append(f.a.catalog, resources.Permission{Code: "system.fixture.read", Name: "Audit", Group: "Audit", Kind: "menu"})
	require(t, f.a.SyncResources(ctx))
	require(t, f.a.SyncResources(ctx))
	after, err := f.b.Current(ctx, user)
	require(t, err)
	if !after.Has(resources.UserRead) || after.Has("system.fixture.read") || after.Policy.Snapshot.Version <= before.Policy.Snapshot.Version {
		t.Fatal("sync changed grants or failed to version")
	}
	root, err := f.b.Current(ctx, f.root)
	require(t, err)
	if !root.Has("system.fixture.read") {
		t.Fatal("super role not complete")
	}
	// Removing a catalog entry deactivates it while preserving grant records for reviewed reintroduction.
	catalog := f.a.catalog
	filtered := []resources.Permission{}
	for _, p := range catalog {
		if p.Code != resources.UserRead {
			filtered = append(filtered, p)
		}
	}
	f.a.catalog = filtered
	require(t, f.a.SyncResources(ctx))
	after, err = f.b.Current(ctx, user)
	require(t, err)
	if after.Has(resources.UserRead) {
		t.Fatal("inactive permission allowed")
	}
	f.a.catalog = catalog
	require(t, f.a.SyncResources(ctx))
	after, err = f.b.Current(ctx, user)
	require(t, err)
	if !after.Has(resources.UserRead) {
		t.Fatal("grant was destroyed by sync")
	}
	// Loaded snapshots never authorize when the authoritative database cannot be checked.
	pool, _ := f.db.DB()
	require(t, pool.Close())
	_, err = f.b.Current(ctx, user)
	code(t, err, codes.Unavailable)
}
