package member

import (
	"context"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
	"github.com/zeromicro/go-zero/core/trace"
)

type ListMemberUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMemberUsersLogic(ctx context.Context, s *svc.ServiceContext) *ListMemberUsersLogic {
	return &ListMemberUsersLogic{ctx, s}
}
func (l *ListMemberUsersLogic) ListMemberUsers(req *types.ListManagementRequest) (*types.MemberUserListResponse, error) {
	reply, err := l.svcCtx.Member.ListUsers(l.ctx, &pb.UserQuery{Page: req.Page, PageSize: req.PageSize, Keyword: req.Keyword, Enabled: req.Enabled})
	if err != nil {
		return nil, err
	}
	items := []types.MemberUserRecord{}
	for _, u := range reply.Items {
		items = append(items, record(u))
	}
	return &types.MemberUserListResponse{Message: "ok", Data: types.MemberUserListData{Items: items, Total: reply.Total, Page: req.Page, PageSize: req.PageSize}, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
