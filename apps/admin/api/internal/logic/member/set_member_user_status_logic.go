package member

import (
	"context"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
	"github.com/zeromicro/go-zero/core/trace"
)

type SetMemberUserStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetMemberUserStatusLogic(ctx context.Context, s *svc.ServiceContext) *SetMemberUserStatusLogic {
	return &SetMemberUserStatusLogic{ctx, s}
}
func (l *SetMemberUserStatusLogic) SetMemberUserStatus(req *types.SetMemberUserStatusRequest) (*types.ActionResponse, error) {
	_, err := l.svcCtx.Member.SetUserStatus(l.ctx, &pb.SetUserStatusRequest{Id: req.Id, Enabled: req.Enabled, Revision: req.Revision})
	if err != nil {
		return nil, err
	}
	return &types.ActionResponse{Message: "ok", Data: types.ActionData{Success: true}, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
