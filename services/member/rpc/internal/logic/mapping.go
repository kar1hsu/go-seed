package logic

import (
	"github.com/kar1hsu/go-seed/services/member/internal/application/account"
	"github.com/kar1hsu/go-seed/services/member/internal/model"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
	"strconv"
	"time"
)

func profile(u *model.User) *pb.Profile {
	return &pb.Profile{Id: strconv.FormatUint(u.ID, 10), Username: u.Username, Nickname: u.Nickname, Revision: account.Revision(u.UpdatedAt), CreatedAt: u.CreatedAt.UTC().Format(time.RFC3339)}
}
func session(s *account.Session) *pb.SessionResponse {
	return &pb.SessionResponse{User: profile(s.User), AccessToken: s.AccessToken, ExpiresAt: s.ExpiresAt.UTC().Format(time.RFC3339), RefreshToken: s.RefreshToken, RefreshExpiresAt: s.RefreshExpiresAt.UTC().Format(time.RFC3339)}
}
func user(u *model.User) *pb.UserRecord {
	r := &pb.UserRecord{Id: strconv.FormatUint(u.ID, 10), Username: u.Username, Nickname: u.Nickname, Enabled: u.Enabled, Revision: account.Revision(u.UpdatedAt), CreatedAt: u.CreatedAt.UTC().Format(time.RFC3339), UpdatedAt: u.UpdatedAt.UTC().Format(time.RFC3339)}
	if u.LastLoginAt != nil {
		r.LastLoginAt = u.LastLoginAt.UTC().Format(time.RFC3339)
	}
	return r
}
