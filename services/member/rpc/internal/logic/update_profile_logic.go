package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/member/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
)

type UpdateProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProfileLogic(ctx context.Context, s *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{ctx, s}
}
func (l *UpdateProfileLogic) UpdateProfile(in *pb.UpdateProfileRequest) (*pb.Profile, error) {
	u, err := l.svcCtx.Account.UpdateProfile(l.ctx, identity.RawToken(l.ctx), in.Nickname, in.Revision)
	if err != nil {
		return nil, err
	}
	return profile(u), nil
}
