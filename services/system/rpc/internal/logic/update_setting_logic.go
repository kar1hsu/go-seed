package logic

import (
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"

	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateSettingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSettingLogic {
	return &UpdateSettingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateSettingLogic) UpdateSetting(in *pb.UpdateSettingRequest) (*pb.AdminActionResponse, error) {
	id, err := management.ID(in.Id)
	if err != nil {
		return nil, err
	}
	input := management.SettingInput{Name: in.Name, Description: in.Description, Enabled: in.Enabled, Value: in.Value}
	err = l.svcCtx.Management.UpdateSetting(l.ctx, identity.RawToken(l.ctx), id, input, in.Revision)
	return &pb.AdminActionResponse{}, err
}
