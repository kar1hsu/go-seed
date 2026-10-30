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

type ListAuditLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAuditLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAuditLogsLogic {
	return &ListAuditLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAuditLogsLogic) ListAuditLogs(req *types.ListAuditLogsRequest) (resp *types.AuditLogListResponse, err error) {

	reply, err := l.svcCtx.System.ListAuditLogs(l.ctx, &system.ListAuditLogsRequest{Page: req.Page, PageSize: req.PageSize, Keyword: req.Keyword, Action: req.Action, ActorId: req.ActorId, TraceId: req.TraceId, From: req.From, To: req.To})
	if err != nil {
		return nil, err
	}
	data := types.AuditLogListData{Items: []types.AuditLogRecord{}, Actions: []types.AuditAction{}, Total: reply.Total, Page: req.Page, PageSize: req.PageSize}
	for _, item := range reply.Items {
		data.Items = append(data.Items, auditRecord(item))
	}
	for _, action := range reply.Actions {
		data.Actions = append(data.Actions, types.AuditAction{Code: action.Code, Name: action.Name})
	}
	return &types.AuditLogListResponse{Message: "ok", Data: data, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
