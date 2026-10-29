// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package auth

import (
	"context"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/authhttp"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Sign in; returns access token and sets HttpOnly refresh cookie
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (*authhttp.Session, error) {
	if len(req.Username) < 3 || len(req.Username) > 32 || len(req.Password) == 0 || len(req.Password) > 72 {
		return nil, xerr.New(codes.InvalidArgument)
	}
	reply, err := l.svcCtx.System.AdminLogin(l.ctx, &pb.AdminLoginRequest{Username: req.Username, Password: req.Password})
	if err != nil {
		return nil, err
	}
	return authhttp.Response(l.ctx, reply)
}
