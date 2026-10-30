package authorization

import (
	"context"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestDisabledRoleIsManageableOnlyWithinExistingAuthority(t *testing.T) {
	role := model.Role{ID: 1, Enabled: false}
	policy := &Policy{Snapshot: &repository.PolicySnapshot{Roles: []model.Role{role}, Permissions: []model.Permission{{Code: "example.read", Active: true}}, Grants: []model.RolePermission{{RoleID: 1, PermissionCode: "example.read"}}, Memberships: []model.AdminRole{{AdminID: 2, RoleID: 1}}}}
	a := &Access{Permissions: []string{"example.read"}, Policy: policy}
	if !a.CanManageRole(role) || a.CanAssignRole(role) {
		t.Fatal("disabled role grant/management confused")
	}
	a.Permissions = nil
	if a.CanManageRole(role) || a.CanManageAdmin(2) {
		t.Fatal("disabled role bypasses authority bounds")
	}
	role.BuiltIn = true
	a.SuperAdmin = true
	if a.CanManageRole(role) {
		t.Fatal("built-in role became editable")
	}
}
func TestWorkerAuthorizerCannotAcceptBearerTokens(t *testing.T) {
	_, err := NewDeferred().Current(context.Background(), nil, "forged")
	if status.Code(err) != codes.Unauthenticated {
		t.Fatal(err)
	}
}
