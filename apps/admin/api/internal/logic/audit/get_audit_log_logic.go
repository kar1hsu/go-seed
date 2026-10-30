// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package audit

import (
	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/system"
	"github.com/zeromicro/go-zero/core/trace"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAuditLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAuditLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAuditLogLogic {
	return &GetAuditLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAuditLogLogic) GetAuditLog(req *types.GetAuditLogRequest) (resp *types.AuditLogDetailResponse, err error) {

	reply, err := l.svcCtx.System.GetAuditLog(l.ctx, &system.GetAuditLogRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}
	data := types.AuditLogDetailData{Record: auditRecord(reply.Record), Changes: []types.AuditChange{}}
	for _, change := range reply.Changes {
		data.Changes = append(data.Changes, types.AuditChange{Field: change.Field, Label: change.Label, Before: append([]string{}, change.Before...), After: append([]string{}, change.After...)})
	}
	return &types.AuditLogDetailResponse{Message: "ok", Data: data, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
