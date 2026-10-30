package member

import (
	"context"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/authhttp"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/svc"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
)

type MemberRefreshLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMemberRefreshLogic(ctx context.Context, s *svc.ServiceContext) *MemberRefreshLogic {
	return &MemberRefreshLogic{ctx, s}
}
func (l *MemberRefreshLogic) MemberRefresh(raw string) (*authhttp.Session, error) {
	reply, err := l.svcCtx.Member.Refresh(l.ctx, &pb.RefreshRequest{RefreshToken: raw})
	if err != nil {
		return nil, err
	}
	return authhttp.Response(l.ctx, reply)
}
