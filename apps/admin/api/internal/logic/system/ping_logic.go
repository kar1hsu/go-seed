// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package system

import (
	"context"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/trace"
)

type PingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Probe API to system-rpc connectivity
func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PingLogic) Ping() (resp *types.PingResponse, err error) {
	reply, err := l.svcCtx.System.Ping(l.ctx, &pb.PingRequest{})
	if err != nil {
		return nil, err
	}
	return &types.PingResponse{Code: 0, Message: "ok", Data: types.PingData{Service: reply.Service, Version: reply.Version, Time: reply.Time}, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
