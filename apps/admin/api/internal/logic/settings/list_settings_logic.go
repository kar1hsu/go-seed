// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package settings

import (
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/trace"

	"context"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSettingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSettingsLogic {
	return &ListSettingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListSettingsLogic) ListSettings(req *types.ListManagementRequest) (resp *types.SettingListResponse, err error) {
	reply, err := l.svcCtx.System.ListSettings(l.ctx, &pb.ListManagementRequest{Page: req.Page, PageSize: req.PageSize, Keyword: req.Keyword, Enabled: req.Enabled})
	if err != nil {
		return nil, err
	}
	items := []types.SettingRecord{}
	for _, item := range reply.Items {
		items = append(items, settingRecord(item))
	}
	return &types.SettingListResponse{Message: "ok", Data: types.SettingListData{Items: items, Total: reply.Total, Page: req.Page, PageSize: req.PageSize}, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
