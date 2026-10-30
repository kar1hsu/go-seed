package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/services/member/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, s *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{ctx, s}
}
func (l *RegisterLogic) Register(in *pb.RegisterRequest) (*pb.Profile, error) {
	u, err := l.svcCtx.Account.Register(l.ctx, in.Username, in.Password, in.Nickname)
	if err != nil {
		return nil, err
	}
	return profile(u), nil
}
