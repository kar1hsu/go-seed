package logic

import (
	"context"

	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"
	"github.com/kar1hsu/go-seed/services/system/internal/audit"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAuditLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAuditLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAuditLogsLogic {
	return &ListAuditLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListAuditLogsLogic) ListAuditLogs(in *pb.ListAuditLogsRequest) (*pb.ListAuditLogsResponse, error) {

	items, total, err := l.svcCtx.Management.ListAuditLogs(l.ctx, identity.RawToken(l.ctx), management.AuditQuery{Page: in.Page, PageSize: in.PageSize, Keyword: in.Keyword, Action: in.Action, ActorID: in.ActorId, TraceID: in.TraceId, From: in.From, To: in.To})
	if err != nil {
		return nil, err
	}
	result := &pb.ListAuditLogsResponse{Items: []*pb.AuditLogRecord{}, Actions: []*pb.AuditAction{}, Total: total}
	for _, item := range items {
		result.Items = append(result.Items, auditRecord(item))
	}
	for _, action := range audit.Actions() {
		result.Actions = append(result.Actions, &pb.AuditAction{Code: action.Code, Name: action.Name})
	}
	return result, nil
}
