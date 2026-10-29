package logic

import (
	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type AdminRefreshLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminRefreshLogic {
	return &AdminRefreshLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdminRefreshLogic) AdminRefresh(in *pb.AdminRefreshRequest) (*pb.AdminAuthResponse, error) {
	result, err := l.svcCtx.Auth.Refresh(l.ctx, in.RefreshToken)
	if err != nil {
		return nil, err
	}
	return authResponse(result), nil
}
