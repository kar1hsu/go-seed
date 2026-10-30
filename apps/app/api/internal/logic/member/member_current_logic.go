package member

import (
	"context"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/types"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
)

type MemberCurrentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMemberCurrentLogic(ctx context.Context, s *svc.ServiceContext) *MemberCurrentLogic {
	return &MemberCurrentLogic{ctx, s}
}
func (l *MemberCurrentLogic) MemberCurrent() (*types.MemberProfileResponse, error) {
	reply, err := l.svcCtx.Member.Current(l.ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}
	return profileResponse(l.ctx, reply), nil
}
