package management

import (
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
)

func administratorRecord(a *pb.AdministratorRecord) types.AdministratorRecord {
	return types.AdministratorRecord{Id: a.Id, Username: a.Username, DisplayName: a.DisplayName, Enabled: a.Enabled, RoleIds: append([]string{}, a.RoleIds...), RoleNames: append([]string{}, a.RoleNames...), Revision: a.Revision, CreatedAt: a.CreatedAt, Manageable: a.Manageable}
}
func roleRecord(r *pb.RoleRecord) types.RoleRecord {
	return types.RoleRecord{Id: r.Id, Code: r.Code, Name: r.Name, Description: r.Description, Enabled: r.Enabled, BuiltIn: r.BuiltIn, Permissions: append([]string{}, r.Permissions...), MemberCount: r.MemberCount, Revision: r.Revision, Manageable: r.Manageable, Assignable: r.Assignable}
}
func managementQuery(r *types.ListManagementRequest) *pb.ListManagementRequest {
	return &pb.ListManagementRequest{Page: r.Page, PageSize: r.PageSize, Keyword: r.Keyword, Enabled: r.Enabled}
}
