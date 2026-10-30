package member

import (
	"context"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/types"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
)

type MemberLogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMemberLogoutLogic(ctx context.Context, s *svc.ServiceContext) *MemberLogoutLogic {
	return &MemberLogoutLogic{ctx, s}
}
func (l *MemberLogoutLogic) MemberLogout(raw string) (*types.MemberActionResponse, error) {
	_, err := l.svcCtx.Member.Logout(l.ctx, &pb.RefreshRequest{RefreshToken: raw})
	return action(l.ctx), err
}
