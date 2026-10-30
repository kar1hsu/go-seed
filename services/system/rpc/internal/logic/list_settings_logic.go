package logic

import (
	"github.com/kar1hsu/go-seed/pkg/identity"

	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSettingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSettingsLogic {
	return &ListSettingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListSettingsLogic) ListSettings(in *pb.ListManagementRequest) (*pb.ListSettingsResponse, error) {
	items, total, err := l.svcCtx.Management.ListSettings(l.ctx, identity.RawToken(l.ctx), managementQuery(in))
	if err != nil {
		return nil, err
	}
	result := &pb.ListSettingsResponse{Items: []*pb.SettingRecord{}, Total: total}
	for _, item := range items {
		result.Items = append(result.Items, settingRecord(item))
	}
	return result, nil
}
