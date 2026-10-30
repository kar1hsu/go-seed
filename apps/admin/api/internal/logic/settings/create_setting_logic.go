// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package settings

import (
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/trace"

	"context"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSettingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSettingLogic {
	return &CreateSettingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSettingLogic) CreateSetting(req *types.CreateSettingRequest) (resp *types.SettingResponse, err error) {
	reply, err := l.svcCtx.System.CreateSetting(l.ctx, &pb.CreateSettingRequest{Name: req.Name, Description: req.Description, Enabled: req.Enabled, Value: req.Value, Code: req.Code, Type: req.Type})
	if err != nil {
		return nil, err
	}
	return &types.SettingResponse{Message: "ok", Data: settingRecord(reply.Record), TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
