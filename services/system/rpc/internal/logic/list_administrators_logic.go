package logic

import (
	"context"

	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListAdministratorsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAdministratorsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAdministratorsLogic {
	return &ListAdministratorsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListAdministratorsLogic) ListAdministrators(in *pb.ListManagementRequest) (*pb.ListAdministratorsResponse, error) {
	items, total, err := l.svcCtx.Management.ListAdministrators(l.ctx, identity.RawToken(l.ctx), managementQuery(in))
	if err != nil {
		return nil, err
	}
	result := &pb.ListAdministratorsResponse{Items: []*pb.AdministratorRecord{}, Total: total}
	for _, item := range items {
		result.Items = append(result.Items, administratorRecord(item))
	}
	return result, nil
}
