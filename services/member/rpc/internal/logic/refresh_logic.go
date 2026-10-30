package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/services/member/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
)

type RefreshLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshLogic(ctx context.Context, s *svc.ServiceContext) *RefreshLogic {
	return &RefreshLogic{ctx, s}
}
func (l *RefreshLogic) Refresh(in *pb.RefreshRequest) (*pb.SessionResponse, error) {
	s, err := l.svcCtx.Account.Refresh(l.ctx, in.RefreshToken)
	if err != nil {
		return nil, err
	}
	return session(s), nil
}
