package logic

import (
	"context"

	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type AdminCurrentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminCurrentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCurrentLogic {
	return &AdminCurrentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdminCurrentLogic) AdminCurrent(in *pb.AdminCurrentRequest) (*pb.AdminCurrentResponse, error) {
	a, err := l.svcCtx.Auth.Current(l.ctx, identity.RawToken(l.ctx))
	if err != nil {
		return nil, err
	}
	return &pb.AdminCurrentResponse{Admin: adminProfile(a)}, nil
}
