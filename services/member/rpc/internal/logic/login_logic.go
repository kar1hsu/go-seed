package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/services/member/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, s *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{ctx, s}
}
func (l *LoginLogic) Login(in *pb.LoginRequest) (*pb.SessionResponse, error) {
	s, err := l.svcCtx.Account.Login(l.ctx, in.Username, in.Password)
	if err != nil {
		return nil, err
	}
	return session(s), nil
}
