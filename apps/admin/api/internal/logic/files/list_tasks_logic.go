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

type ListTasksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTasksLogic {
	return &ListTasksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTasksLogic) ListTasks(req *types.ListFilesTasksRequest) (resp *types.TasksResponse, err error) {
	reply, err := l.svcCtx.System.ListTasks(l.ctx, &pb.ListFilesTasksRequest{Page: req.Page, PageSize: req.PageSize, Keyword: req.Keyword, Status: req.Status})
	if err != nil {
		return nil, err
	}
	result := types.TasksData{Items: []types.TaskRecord{}, Total: reply.Total, Page: req.Page, PageSize: req.PageSize}
	for _, item := range reply.Items {
		result.Items = append(result.Items, taskRecord(item))
	}
	return &types.TasksResponse{Message: "ok", Data: result, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
