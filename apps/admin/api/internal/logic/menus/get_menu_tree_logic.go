// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package menus

import (
	"context"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/menuhttp"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Read the editable navigation tree
func NewGetMenuTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuTreeLogic {
	return &GetMenuTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMenuTreeLogic) GetMenuTree() (resp *types.MenuTreeResponse, err error) {
	reply, err := l.svcCtx.System.GetMenuTree(l.ctx, &pb.GetMenuTreeRequest{})
	return menuhttp.Response(l.ctx, reply, err)
}
