package member

import (
	"context"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
	"github.com/zeromicro/go-zero/core/trace"
)

type GetMemberUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMemberUserLogic(ctx context.Context, s *svc.ServiceContext) *GetMemberUserLogic {
	return &GetMemberUserLogic{ctx, s}
}
func (l *GetMemberUserLogic) GetMemberUser(req *types.MemberUserID) (*types.MemberUserResponse, error) {
	reply, err := l.svcCtx.Member.GetUser(l.ctx, &pb.UserID{Id: req.Id})
	if err != nil {
		return nil, err
	}
	events := []types.MemberUserEvent{}
	for _, e := range reply.RecentEvents {
		events = append(events, types.MemberUserEvent{Id: e.Id, Action: e.Action, ActorKind: e.ActorKind, ActorId: e.ActorId, CreatedAt: e.CreatedAt, TraceId: e.TraceId})
	}
	return &types.MemberUserResponse{Message: "ok", Data: types.MemberUserDetail{User: record(reply.User), RecentEvents: events}, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
