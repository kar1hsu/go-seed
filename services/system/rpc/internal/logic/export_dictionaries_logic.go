package logic

import (
	"github.com/kar1hsu/go-seed/pkg/identity"

	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExportDictionariesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExportDictionariesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExportDictionariesLogic {
	return &ExportDictionariesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ExportDictionariesLogic) ExportDictionaries(in *pb.ExportDictionariesRequest) (*pb.TaskResponse, error) {
	item, err := l.svcCtx.Management.ExportDictionaries(l.ctx, identity.RawToken(l.ctx), in.RequestKey)
	if err != nil {
		return nil, err
	}
	return &pb.TaskResponse{Record: taskRecord(*item)}, nil
}
