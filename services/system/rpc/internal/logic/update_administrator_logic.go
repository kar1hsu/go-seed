package logic

import (
	"context"

	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAdministratorLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAdministratorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAdministratorLogic {
	return &UpdateAdministratorLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateAdministratorLogic) UpdateAdministrator(in *pb.UpdateAdministratorRequest) (*pb.AdminActionResponse, error) {
	id, err := management.ID(in.Id)
	if err != nil {
		return nil, err
	}
	err = l.svcCtx.Management.UpdateAdministrator(l.ctx, identity.RawToken(l.ctx), id, in.DisplayName, in.Enabled, in.Revision)
	return &pb.AdminActionResponse{}, err
}
