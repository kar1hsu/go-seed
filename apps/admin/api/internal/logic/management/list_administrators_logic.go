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

type ListAdministratorsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAdministratorsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAdministratorsLogic {
	return &ListAdministratorsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAdministratorsLogic) ListAdministrators(req *types.ListManagementRequest) (resp *types.AdministratorListResponse, err error) {
	reply, err := l.svcCtx.System.ListAdministrators(l.ctx, managementQuery(req))
	if err != nil {
		return nil, err
	}
	items := []types.AdministratorRecord{}
	for _, item := range reply.Items {
		items = append(items, administratorRecord(item))
	}
	return &types.AdministratorListResponse{Message: "ok", Data: types.AdministratorListData{Items: items, Total: reply.Total, Page: req.Page, PageSize: req.PageSize}, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
