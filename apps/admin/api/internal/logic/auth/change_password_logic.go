// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package auth

import (
	"context"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/trace"
	"google.golang.org/grpc/codes"
)

type ChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Change password and revoke every administrator session
func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangePasswordLogic) ChangePassword(req *types.ChangePasswordRequest) (*types.ActionResponse, error) {
	if len(req.CurrentPassword) == 0 || len(req.CurrentPassword) > 72 || len(req.NewPassword) < 12 || len(req.NewPassword) > 72 {
		return nil, xerr.New(codes.InvalidArgument)
	}
	_, err := l.svcCtx.System.AdminChangePassword(l.ctx, &pb.AdminPasswordRequest{CurrentPassword: req.CurrentPassword, NewPassword: req.NewPassword})
	if err != nil {
		return nil, err
	}
	return &types.ActionResponse{Code: 0, Message: "ok", Data: types.ActionData{Success: true}, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
