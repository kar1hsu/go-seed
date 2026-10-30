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

type DeleteSettingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSettingLogic {
	return &DeleteSettingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteSettingLogic) DeleteSetting(req *types.DeleteSettingResourceRequest) (resp *types.ActionResponse, err error) {
	_, err = l.svcCtx.System.DeleteSetting(l.ctx, &pb.DeleteSettingResourceRequest{Id: req.Id, Revision: req.Revision})
	if err != nil {
		return nil, err
	}
	return &types.ActionResponse{Message: "ok", Data: types.ActionData{Success: true}, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
