package logic

import (
	"strconv"
	"time"

	"github.com/kar1hsu/go-seed/services/system/internal/application/auth"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
)

func adminProfile(a *model.Administrator) *pb.AdminProfile {
	return &pb.AdminProfile{Id: strconv.FormatUint(a.ID, 10), Username: a.Username, DisplayName: a.DisplayName}
}
func authResponse(s *auth.Session) *pb.AdminAuthResponse {
	return &pb.AdminAuthResponse{
		AccessToken: s.AccessToken, ExpiresAt: s.ExpiresAt.Format(time.RFC3339), Admin: adminProfile(s.Admin),
		RefreshToken: s.RefreshToken, RefreshExpiresAt: s.RefreshExpiresAt.Format(time.RFC3339),
	}
}
