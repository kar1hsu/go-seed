package logic

import (
	"context"

	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListPermissionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPermissionsLogic {
	return &ListPermissionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListPermissionsLogic) ListPermissions(in *pb.ListPermissionsRequest) (*pb.ListPermissionsResponse, error) {
	items, err := l.svcCtx.Management.Permissions(l.ctx, identity.RawToken(l.ctx))
	if err != nil {
		return nil, err
	}
	result := &pb.ListPermissionsResponse{Items: []*pb.PermissionRecord{}}
	for _, p := range items {
		result.Items = append(result.Items, &pb.PermissionRecord{Code: p.Code, Name: p.Name, Group: p.Group, Kind: p.Kind, Requires: p.Requires})
	}
	return result, nil
}
