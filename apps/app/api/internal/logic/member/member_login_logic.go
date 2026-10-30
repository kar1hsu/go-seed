package member

import (
	"context"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/authhttp"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/types"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
	"google.golang.org/grpc/codes"
)

type MemberLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMemberLoginLogic(ctx context.Context, s *svc.ServiceContext) *MemberLoginLogic {
	return &MemberLoginLogic{ctx, s}
}
func (l *MemberLoginLogic) MemberLogin(req *types.MemberLoginRequest) (*authhttp.Session, error) {
	if len(req.Username) < 3 || len(req.Username) > 32 || len(req.Password) == 0 || len(req.Password) > 72 {
		return nil, xerr.New(codes.InvalidArgument)
	}
	reply, err := l.svcCtx.Member.Login(l.ctx, &pb.LoginRequest{Username: req.Username, Password: req.Password})
	if err != nil {
		return nil, err
	}
	return authhttp.Response(l.ctx, reply)
}
