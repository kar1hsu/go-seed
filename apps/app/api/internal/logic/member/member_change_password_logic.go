package member

import (
	"context"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/types"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
)

type MemberChangePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMemberChangePasswordLogic(ctx context.Context, s *svc.ServiceContext) *MemberChangePasswordLogic {
	return &MemberChangePasswordLogic{ctx, s}
}
func (l *MemberChangePasswordLogic) MemberChangePassword(req *types.MemberPasswordRequest) (*types.MemberActionResponse, error) {
	_, err := l.svcCtx.Member.ChangePassword(l.ctx, &pb.ChangePasswordRequest{CurrentPassword: req.CurrentPassword, NewPassword: req.NewPassword})
	return action(l.ctx), err
}
