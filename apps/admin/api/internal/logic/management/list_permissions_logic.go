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

type ListPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPermissionsLogic {
	return &ListPermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPermissionsLogic) ListPermissions() (resp *types.PermissionListResponse, err error) {
	reply, err := l.svcCtx.System.ListPermissions(l.ctx, &pb.ListPermissionsRequest{})
	if err != nil {
		return nil, err
	}
	items := []types.PermissionRecord{}
	for _, p := range reply.Items {
		items = append(items, types.PermissionRecord{Code: p.Code, Name: p.Name, Group: p.Group, Kind: p.Kind, Requires: append([]string{}, p.Requires...)})
	}
	return &types.PermissionListResponse{Message: "ok", Data: types.PermissionListData{Items: items}, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
