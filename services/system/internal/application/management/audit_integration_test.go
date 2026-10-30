//go:build integration

package management

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kar1hsu/go-seed/services/system/internal/audit"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/resources"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

func TestAuditLifecycleAndQueries(t *testing.T) {
	f := setup(t)
	traceID, err := trace.TraceIDFromHex("1234567890abcdef1234567890abcdef")
	require(t, err)
	spanID, err := trace.SpanIDFromHex("1234567890abcdef")
	require(t, err)
	requestContext := trace.ContextWithSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID}))
	user, err := f.a.CreateAdministrator(requestContext, f.root, "audit_user", "审计用户", password)
	require(t, err)
	require(t, f.a.UpdateAdministrator(ctx, f.root, user.ID, "已编辑用户", false, Revision(user.UpdatedAt)))
	r, err := f.a.CreateRole(ctx, f.root, "audit_role", "100%操作_角色", "原始说明")
	require(t, err)
	require(t, f.a.UpdateRole(ctx, f.root, r.ID, r.Name, "修改说明", true, Revision(r.UpdatedAt)))
	require(t, f.a.GrantRole(ctx, f.root, r.ID, []string{resources.UserRead}, f.roleRevision(t, r.ID)))
	f.assign(t, user.ID, r.ID)
	f.assign(t, user.ID)
	// Retired catalog entries remain part of the persisted grant history even after they cease granting access.
	for i, p := range f.a.catalog {
		if p.Code == resources.UserRead {
			f.a.catalog = append(f.a.catalog[:i], f.a.catalog[i+1:]...)
			break
		}
	}
	require(t, f.a.SyncResources(ctx))
	require(t, f.a.DeleteRole(ctx, f.root, r.ID, f.roleRevision(t, r.ID)))
	q := AuditQuery{Page: 1, PageSize: 100}
	rows, total, err := f.b.ListAuditLogs(ctx, f.root, q)
	require(t, err)
	if total != 8 || len(rows) != 8 {
		t.Fatalf("expected eight committed mutations, got %d/%d", total, len(rows))
	}
	seen := map[string]bool{}
	for i, row := range rows {
		if i > 0 && row.ID >= rows[i-1].ID {
			t.Fatal("unstable ordering")
		}
		if row.ChangesJSON != "" {
			t.Fatal("list exposed change payload")
		}
		if row.ActorID != f.rootID || row.ActorUsername != "test_root" || row.Source != "admin" {
			t.Fatal("untrusted actor")
		}
		seen[row.Action] = true
		detail, err := f.a.GetAuditLog(ctx, f.root, row.ID)
		require(t, err)
		if strings.Contains(detail.Record.ChangesJSON, password) || strings.Contains(detail.Record.ChangesJSON, "password_hash") || strings.Contains(detail.Record.ChangesJSON, "$2a$") {
			t.Fatal("credential in audit")
		}
	}
	for _, action := range []string{resources.UserCreate, resources.UserUpdate, resources.UserAssignRole, resources.RoleCreate, resources.RoleUpdate, resources.RoleGrant, resources.RoleDelete} {
		if !seen[action] {
			t.Fatalf("missing audit action %s", action)
		}
	}
	deleted, err := f.a.GetAuditLog(ctx, f.root, rows[0].ID)
	require(t, err)
	if deleted.Record.TargetKey != "audit_role" || deleted.Record.TargetName != "100%操作_角色" || !strings.Contains(deleted.Record.ChangesJSON, resources.UserRead) {
		t.Fatal("deleted role snapshot lost")
	}
	q.PageSize = 2
	first, _, err := f.b.ListAuditLogs(ctx, f.root, q)
	require(t, err)
	q.Page = 2
	second, _, err := f.b.ListAuditLogs(ctx, f.root, q)
	require(t, err)
	if len(first) != 2 || len(second) != 2 || first[1].ID <= second[0].ID {
		t.Fatal("pagination overlap")
	}
	q = AuditQuery{Page: 1, PageSize: 100, Keyword: "%操作_"}
	_, total, err = f.b.ListAuditLogs(ctx, f.root, q)
	require(t, err)
	if total != 4 {
		t.Fatalf("literal keyword matched %d, want 4", total)
	}
	q.Keyword = "not_found_%"
	_, total, err = f.b.ListAuditLogs(ctx, f.root, q)
	require(t, err)
	if total != 0 {
		t.Fatal("wildcard keyword matched other targets")
	}
	q = AuditQuery{Page: 1, PageSize: 20, TraceID: traceID.String(), ActorID: strconv.FormatUint(f.rootID, 10), Action: resources.UserCreate}
	rows, total, err = f.b.ListAuditLogs(ctx, f.root, q)
	require(t, err)
	if total != 1 {
		t.Fatal("trace/action/actor filter failed")
	}
	q.From, q.To = Revision(rows[0].CreatedAt), Revision(rows[0].CreatedAt.Add(time.Microsecond))
	_, total, err = f.b.ListAuditLogs(ctx, f.root, q)
	require(t, err)
	if total != 1 {
		t.Fatal("inclusive from excluded event")
	}
	q.From, q.To = "", Revision(rows[0].CreatedAt)
	_, total, err = f.b.ListAuditLogs(ctx, f.root, q)
	require(t, err)
	if total != 0 {
		t.Fatal("exclusive to included event")
	}
	_, err = f.b.GetAuditLog(ctx, f.root, 999999)
	code(t, err, codes.NotFound)
}

func TestAuditAccessRevocationAndFailedMutations(t *testing.T) {
	f := setup(t)
	id, reader := f.user(t, "audit_reader")
	q := AuditQuery{Page: 1, PageSize: 20}
	rows, _, err := f.a.ListAuditLogs(ctx, f.root, q)
	require(t, err)
	_, _, err = f.b.ListAuditLogs(ctx, reader, q)
	code(t, err, codes.PermissionDenied)
	_, err = f.b.GetAuditLog(ctx, reader, rows[0].ID)
	code(t, err, codes.PermissionDenied)
	roleID := f.role(t, "auditor", resources.AuditRead)
	f.assign(t, id, roleID)
	_, _, err = f.b.ListAuditLogs(ctx, reader, q)
	require(t, err)
	_, err = f.b.GetAuditLog(ctx, reader, rows[0].ID)
	require(t, err)
	_, before, err := f.a.ListAuditLogs(ctx, f.root, q)
	require(t, err)
	_, err = f.b.CreateRole(ctx, reader, "forbidden", "forbidden", "")
	code(t, err, codes.PermissionDenied)
	code(t, f.a.UpdateAdministrator(ctx, f.root, id, "stale", true, "invalid-revision"), codes.Aborted)
	_, after, err := f.a.ListAuditLogs(ctx, f.root, q)
	require(t, err)
	if before != after {
		t.Fatal("failed mutation recorded as success")
	}
	require(t, f.a.GrantRole(ctx, f.root, roleID, []string{}, f.roleRevision(t, roleID)))
	_, _, err = f.b.ListAuditLogs(ctx, reader, q)
	code(t, err, codes.PermissionDenied)
	_, err = f.b.GetAuditLog(ctx, reader, rows[0].ID)
	code(t, err, codes.PermissionDenied)
}

func TestAuditFailureRollsBackMutationAndPolicy(t *testing.T) {
	f := setup(t)
	r, err := f.a.CreateRole(ctx, f.root, "atomic_role", "Original", "")
	require(t, err)
	userID, userToken := f.user(t, "atomic_user")
	before, err := repository.NewRBAC(f.db).Version(ctx)
	require(t, err)
	callback := "test:reject_audit_insert"
	require(t, f.db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "audit_logs" {
			tx.AddError(errors.New("audit store unavailable"))
		}
	}))
	t.Cleanup(func() { f.db.Callback().Create().Remove(callback) })
	code(t, f.a.UpdateRole(ctx, f.root, r.ID, "Should roll back", "changed", false, Revision(r.UpdatedAt)), codes.Unavailable)
	_, err = f.a.CreateRole(ctx, f.root, "rolled_back", "Rollback", "")
	code(t, err, codes.Unavailable)
	var actual model.Role
	require(t, f.db.First(&actual, r.ID).Error)
	if actual.Name != r.Name || !actual.Enabled || !actual.UpdatedAt.Equal(r.UpdatedAt) {
		t.Fatal("business change committed without audit")
	}
	var count int64
	require(t, f.db.Model(&model.Role{}).Where("code = ?", "rolled_back").Count(&count).Error)
	if count != 0 {
		t.Fatal("create committed without audit")
	}
	after, err := repository.NewRBAC(f.db).Version(ctx)
	require(t, err)
	if before != after {
		t.Fatal("policy version advanced on rolled back mutation")
	}
	code(t, f.auth.ChangePassword(ctx, userToken, password, "new-atomic-password"), codes.Unavailable)
	current, err := f.auth.Current(ctx, userToken)
	require(t, err)
	if current.ID != userID {
		t.Fatal("old session revoked despite rollback")
	}
	_, err = f.auth.Login(ctx, "atomic_user", password)
	require(t, err)
}

func TestAuditPasswordAndCommandRedaction(t *testing.T) {
	f := setup(t)
	id, token := f.user(t, "sensitive_user")
	newPassword := "another-integration-password"
	require(t, f.auth.ChangePassword(ctx, token, password, newPassword))
	require(t, f.auth.Disable(ctx, "sensitive_user"))
	q := AuditQuery{Page: 1, PageSize: 20, Action: audit.PasswordChange}
	rows, total, err := f.a.ListAuditLogs(ctx, f.root, q)
	require(t, err)
	if total != 1 || rows[0].ActorID != id {
		t.Fatal("missing password action")
	}
	detail, err := f.a.GetAuditLog(ctx, f.root, rows[0].ID)
	require(t, err)
	encoded, err := json.Marshal(detail)
	require(t, err)
	for _, secret := range []string{password, newPassword, token, "$2a$"} {
		if strings.Contains(string(encoded), secret) {
			t.Fatal("audit leaked credential")
		}
	}
	q.Action, q.ActorID = audit.CommandDisable, "0"
	rows, total, err = f.a.ListAuditLogs(ctx, f.root, q)
	require(t, err)
	if total != 1 || rows[0].Source != "command" || rows[0].TargetID != id {
		t.Fatal("command origin not distinguished")
	}
}
