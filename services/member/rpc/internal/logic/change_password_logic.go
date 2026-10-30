package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/member/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
)

type ChangePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangePasswordLogic(ctx context.Context, s *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{ctx, s}
}
func (l *ChangePasswordLogic) ChangePassword(in *pb.ChangePasswordRequest) (*pb.Empty, error) {
	err := l.svcCtx.Account.ChangePassword(l.ctx, identity.RawToken(l.ctx), in.CurrentPassword, in.NewPassword)
	return &pb.Empty{}, err
}
