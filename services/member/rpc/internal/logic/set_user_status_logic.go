package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/services/member/internal/application/management"
	"github.com/kar1hsu/go-seed/services/member/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
)

type SetUserStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetUserStatusLogic(ctx context.Context, s *svc.ServiceContext) *SetUserStatusLogic {
	return &SetUserStatusLogic{ctx, s}
}
func (l *SetUserStatusLogic) SetUserStatus(in *pb.SetUserStatusRequest) (*pb.Empty, error) {
	id, err := management.ID(in.Id)
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, l.svcCtx.Management.SetStatus(l.ctx, id, in.Enabled, in.Revision)
}
