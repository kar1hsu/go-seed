package member

import (
	"context"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/types"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
)

type UpdateMemberProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateMemberProfileLogic(ctx context.Context, s *svc.ServiceContext) *UpdateMemberProfileLogic {
	return &UpdateMemberProfileLogic{ctx, s}
}
func (l *UpdateMemberProfileLogic) UpdateMemberProfile(req *types.UpdateMemberProfileRequest) (*types.MemberProfileResponse, error) {
	reply, err := l.svcCtx.Member.UpdateProfile(l.ctx, &pb.UpdateProfileRequest{Nickname: req.Nickname, Revision: req.Revision})
	if err != nil {
		return nil, err
	}
	return profileResponse(l.ctx, reply), nil
}
