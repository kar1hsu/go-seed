package member

import (
	"context"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/types"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
	"google.golang.org/grpc/codes"
)

type RegisterMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterMemberLogic(ctx context.Context, s *svc.ServiceContext) *RegisterMemberLogic {
	return &RegisterMemberLogic{ctx, s}
}
func (l *RegisterMemberLogic) RegisterMember(req *types.RegisterMemberRequest) (*types.MemberProfileResponse, error) {
	if len(req.Username) < 3 || len(req.Username) > 32 || len(req.Password) < 12 || len(req.Password) > 72 || len([]rune(req.Nickname)) > 64 {
		return nil, xerr.New(codes.InvalidArgument)
	}
	reply, err := l.svcCtx.Member.Register(l.ctx, &pb.RegisterRequest{Username: req.Username, Password: req.Password, Nickname: req.Nickname})
	if err != nil {
		return nil, err
	}
	return profileResponse(l.ctx, reply), nil
}
