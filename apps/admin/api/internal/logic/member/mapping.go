package member

import (
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
)

func record(u *pb.UserRecord) types.MemberUserRecord {
	return types.MemberUserRecord{Id: u.Id, Username: u.Username, Nickname: u.Nickname, Enabled: u.Enabled, Revision: u.Revision, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt, LastLoginAt: u.LastLoginAt}
}
