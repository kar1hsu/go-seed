// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package management

import (
	"context"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/trace"
)

type ListRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRolesLogic {
	return &ListRolesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListRolesLogic) ListRoles(req *types.ListManagementRequest) (resp *types.RoleListResponse, err error) {
	reply, err := l.svcCtx.System.ListRoles(l.ctx, managementQuery(req))
	if err != nil {
		return nil, err
	}
	items := []types.RoleRecord{}
	for _, item := range reply.Items {
		items = append(items, roleRecord(item))
	}
	return &types.RoleListResponse{Message: "ok", Data: types.RoleListData{Items: items, Total: reply.Total, Page: req.Page, PageSize: req.PageSize}, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
