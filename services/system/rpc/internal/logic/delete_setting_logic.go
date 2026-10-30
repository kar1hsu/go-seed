package logic

import (
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"

	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSettingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSettingLogic {
	return &DeleteSettingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteSettingLogic) DeleteSetting(in *pb.DeleteSettingResourceRequest) (*pb.AdminActionResponse, error) {
	id, err := management.ID(in.Id)
	if err != nil {
		return nil, err
	}
	err = l.svcCtx.Management.DeleteSetting(l.ctx, identity.RawToken(l.ctx), id, in.Revision)
	return &pb.AdminActionResponse{}, err
}
