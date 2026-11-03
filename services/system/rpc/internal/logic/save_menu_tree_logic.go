package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/pkg/identity"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveMenuTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveMenuTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveMenuTreeLogic {
	return &SaveMenuTreeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SaveMenuTreeLogic) SaveMenuTree(in *pb.MenuTree) (*pb.MenuTree, error) {
	layout, err := l.svcCtx.Management.SaveMenuTree(l.ctx, identity.RawToken(l.ctx), in.GetRevision(), menuGroups(in.GetGroups()))
	if err != nil {
		return nil, err
	}
	return menuTree(layout), nil
}
