// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package auth

import (
	"context"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/authhttp"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/trace"
)

type CurrentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Read the current administrator; checks live session state
func NewCurrentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CurrentLogic {
	return &CurrentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CurrentLogic) Current() (*types.CurrentResponse, error) {
	reply, err := l.svcCtx.System.AdminCurrent(l.ctx, &pb.AdminCurrentRequest{})
	if err != nil {
		return nil, err
	}
	return &types.CurrentResponse{Code: 0, Message: "ok", Data: types.CurrentData{Admin: authhttp.Profile(reply.Admin)}, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
