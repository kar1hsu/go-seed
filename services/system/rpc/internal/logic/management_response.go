package logic

import (
	"strconv"

	"github.com/kar1hsu/go-seed/services/system/internal/application/management"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
)

func administratorRecord(a management.Administrator) *pb.AdministratorRecord {
	ids, names := []string{}, []string{}
	for _, role := range a.Roles {
		ids = append(ids, strconv.FormatUint(role.ID, 10))
		names = append(names, role.Name)
	}
	return &pb.AdministratorRecord{Id: strconv.FormatUint(a.ID, 10), Username: a.Username, DisplayName: a.DisplayName, Enabled: a.Enabled, RoleIds: ids, RoleNames: names, Revision: management.Revision(a.UpdatedAt), CreatedAt: management.Revision(a.CreatedAt), Manageable: a.Manageable}
}
func roleRecord(r management.Role) *pb.RoleRecord {
	return &pb.RoleRecord{Id: strconv.FormatUint(r.ID, 10), Code: r.Code, Name: r.Name, Description: r.Description, Enabled: r.Enabled, BuiltIn: r.BuiltIn, Permissions: r.Permissions, MemberCount: r.MemberCount, Revision: management.Revision(r.UpdatedAt), Manageable: r.Manageable, Assignable: r.Assignable}
}
func managementQuery(r *pb.ListManagementRequest) management.Query {
	return management.Query{Page: r.Page, PageSize: r.PageSize, Keyword: r.Keyword, Enabled: r.Enabled}
}
