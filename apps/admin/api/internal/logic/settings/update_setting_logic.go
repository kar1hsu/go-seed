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

type UpdateSettingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSettingLogic {
	return &UpdateSettingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateSettingLogic) UpdateSetting(req *types.UpdateSettingRequest) (resp *types.ActionResponse, err error) {
	_, err = l.svcCtx.System.UpdateSetting(l.ctx, &pb.UpdateSettingRequest{Name: req.Name, Description: req.Description, Enabled: req.Enabled, Value: req.Value, Id: req.Id, Revision: req.Revision})
	if err != nil {
		return nil, err
	}
	return &types.ActionResponse{Message: "ok", Data: types.ActionData{Success: true}, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
