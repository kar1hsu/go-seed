// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package management

import (
	"context"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/trace"
)

type UpdateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateRoleLogic) UpdateRole(req *types.UpdateRoleRequest) (resp *types.ActionResponse, err error) {
	_, err = l.svcCtx.System.UpdateRole(l.ctx, &pb.UpdateRoleRequest{Id: req.Id, Name: req.Name, Description: req.Description, Enabled: req.Enabled, Revision: req.Revision})
	if err != nil {
		return nil, err
	}
	return &types.ActionResponse{Message: "ok", Data: types.ActionData{Success: true}, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
