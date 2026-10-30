package logic

import (
	"context"

	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAuditLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAuditLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAuditLogLogic {
	return &GetAuditLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAuditLogLogic) GetAuditLog(in *pb.GetAuditLogRequest) (*pb.GetAuditLogResponse, error) {

	id, err := management.ID(in.Id)
	if err != nil {
		return nil, err
	}
	detail, err := l.svcCtx.Management.GetAuditLog(l.ctx, identity.RawToken(l.ctx), id)
	if err != nil {
		return nil, err
	}
	result := &pb.GetAuditLogResponse{Record: auditRecord(detail.Record), Changes: []*pb.AuditChange{}}
	for _, change := range detail.Changes {
		result.Changes = append(result.Changes, &pb.AuditChange{Field: change.Field, Label: change.Label, Before: change.Before, After: change.After})
	}
	return result, nil
}
