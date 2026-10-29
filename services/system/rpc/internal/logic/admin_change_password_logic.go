package logic

import (
	"context"

	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type AdminChangePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminChangePasswordLogic {
	return &AdminChangePasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdminChangePasswordLogic) AdminChangePassword(in *pb.AdminPasswordRequest) (*pb.AdminActionResponse, error) {
	if err := l.svcCtx.Auth.ChangePassword(l.ctx, identity.RawToken(l.ctx), in.CurrentPassword, in.NewPassword); err != nil {
		return nil, err
	}
	return &pb.AdminActionResponse{}, nil
}
