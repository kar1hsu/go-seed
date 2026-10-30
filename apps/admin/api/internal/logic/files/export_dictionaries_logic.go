// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package files

import (
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/trace"

	"context"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExportDictionariesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExportDictionariesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExportDictionariesLogic {
	return &ExportDictionariesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExportDictionariesLogic) ExportDictionaries(req *types.ExportDictionariesRequest) (resp *types.TaskResponse, err error) {
	reply, err := l.svcCtx.System.ExportDictionaries(l.ctx, &pb.ExportDictionariesRequest{RequestKey: req.RequestKey})
	if err != nil {
		return nil, err
	}
	result := types.TaskData{Record: taskRecord(reply.Record)}
	return &types.TaskResponse{Message: "ok", Data: result, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
