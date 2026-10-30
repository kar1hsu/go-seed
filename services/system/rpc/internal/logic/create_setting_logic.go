package logic

import (
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"

	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSettingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSettingLogic {
	return &CreateSettingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateSettingLogic) CreateSetting(in *pb.CreateSettingRequest) (*pb.SettingResponse, error) {
	input := management.SettingInput{Name: in.Name, Description: in.Description, Enabled: in.Enabled, Value: in.Value}
	item, err := l.svcCtx.Management.CreateSetting(l.ctx, identity.RawToken(l.ctx), in.Code, in.Type, input)
	if err != nil {
		return nil, err
	}
	return &pb.SettingResponse{Record: settingRecord(*item)}, nil
}
