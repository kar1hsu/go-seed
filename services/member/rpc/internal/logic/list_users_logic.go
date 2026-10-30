package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/services/member/internal/application/management"
	"github.com/kar1hsu/go-seed/services/member/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
)

type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUsersLogic(ctx context.Context, s *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{ctx, s}
}
func (l *ListUsersLogic) ListUsers(in *pb.UserQuery) (*pb.UserList, error) {
	items, total, err := l.svcCtx.Management.List(l.ctx, management.Query{Page: in.Page, PageSize: in.PageSize, Keyword: in.Keyword, Enabled: in.Enabled})
	if err != nil {
		return nil, err
	}
	result := &pb.UserList{Items: []*pb.UserRecord{}, Total: total}
	for i := range items {
		result.Items = append(result.Items, user(&items[i]))
	}
	return result, nil
}
