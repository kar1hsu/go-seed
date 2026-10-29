package logic

import (
	"context"

	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRolesLogic {
	return &ListRolesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListRolesLogic) ListRoles(in *pb.ListManagementRequest) (*pb.ListRolesResponse, error) {
	items, total, err := l.svcCtx.Management.ListRoles(l.ctx, identity.RawToken(l.ctx), managementQuery(in))
	if err != nil {
		return nil, err
	}
	result := &pb.ListRolesResponse{Items: []*pb.RoleRecord{}, Total: total}
	for _, item := range items {
		result.Items = append(result.Items, roleRecord(item))
	}
	return result, nil
}
