package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/services/member/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
)

type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogoutLogic(ctx context.Context, s *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{ctx, s}
}
func (l *LogoutLogic) Logout(in *pb.RefreshRequest) (*pb.Empty, error) {
	err := l.svcCtx.Account.Logout(l.ctx, in.RefreshToken)
	return &pb.Empty{}, err
}
