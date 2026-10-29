package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"google.golang.org/grpc/codes"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckReadyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckReadyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckReadyLogic {
	return &CheckReadyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Checks MySQL and Redis with the caller's deadline.
func (l *CheckReadyLogic) CheckReady(in *pb.CheckReadyRequest) (*pb.CheckReadyResponse, error) {
	if err := l.svcCtx.Health.Ready(l.ctx); err != nil {
		l.Errorf("readiness failed: %v", err)
		return nil, xerr.New(codes.Unavailable)
	}

	return &pb.CheckReadyResponse{}, nil
}
