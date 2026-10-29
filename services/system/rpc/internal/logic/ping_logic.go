package logic

import (
	"context"
	"time"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Non-sensitive connectivity probe; independent of database readiness.
func (l *PingLogic) Ping(in *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{Service: l.svcCtx.Config.Name, Version: l.svcCtx.Config.Version, Time: time.Now().UTC().Format(time.RFC3339)}, nil
}
