// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package menus

import (
	"context"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/menuhttp"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveMenuTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Save the ordered navigation tree with revision checking
func NewSaveMenuTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveMenuTreeLogic {
	return &SaveMenuTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveMenuTreeLogic) SaveMenuTree(req *types.MenuTree) (resp *types.MenuTreeResponse, err error) {
	reply, err := l.svcCtx.System.SaveMenuTree(l.ctx, menuhttp.Request(req))
	return menuhttp.Response(l.ctx, reply, err)
}
