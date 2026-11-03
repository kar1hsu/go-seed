package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/pkg/identity"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMenuTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuTreeLogic {
	return &GetMenuTreeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMenuTreeLogic) GetMenuTree(in *pb.GetMenuTreeRequest) (*pb.MenuTree, error) {
	layout, err := l.svcCtx.Management.GetMenuTree(l.ctx, identity.RawToken(l.ctx))
	if err != nil {
		return nil, err
	}
	return menuTree(layout), nil
}
