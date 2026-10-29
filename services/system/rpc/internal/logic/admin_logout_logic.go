package logic

import (
	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type AdminLogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLogoutLogic {
	return &AdminLogoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdminLogoutLogic) AdminLogout(in *pb.AdminLogoutRequest) (*pb.AdminActionResponse, error) {
	if err := l.svcCtx.Auth.Logout(l.ctx, in.RefreshToken); err != nil {
		return nil, err
	}
	return &pb.AdminActionResponse{}, nil
}
