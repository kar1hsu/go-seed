package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/member/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
)

type CurrentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCurrentLogic(ctx context.Context, s *svc.ServiceContext) *CurrentLogic {
	return &CurrentLogic{ctx, s}
}
func (l *CurrentLogic) Current(in *pb.Empty) (*pb.Profile, error) {
	u, err := l.svcCtx.Account.Current(l.ctx, identity.RawToken(l.ctx))
	if err != nil {
		return nil, err
	}
	return profile(u), nil
}
