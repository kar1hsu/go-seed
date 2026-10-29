package logic

import (
	"context"

	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type GrantRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGrantRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GrantRoleLogic {
	return &GrantRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GrantRoleLogic) GrantRole(in *pb.GrantRoleRequest) (*pb.AdminActionResponse, error) {
	id, err := management.ID(in.Id)
	if err != nil {
		return nil, err
	}
	err = l.svcCtx.Management.GrantRole(l.ctx, identity.RawToken(l.ctx), id, in.Permissions, in.Revision)
	return &pb.AdminActionResponse{}, err
}
