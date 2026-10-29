// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package auth

import (
	"context"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/authhttp"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Rotate HttpOnly refresh cookie and access token
func NewRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshLogic {
	return &RefreshLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshLogic) Refresh(raw string) (*authhttp.Session, error) {
	reply, err := l.svcCtx.System.AdminRefresh(l.ctx, &pb.AdminRefreshRequest{RefreshToken: raw})
	if err != nil {
		return nil, err
	}
	return authhttp.Response(l.ctx, reply)
}
