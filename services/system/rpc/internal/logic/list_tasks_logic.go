package logic

import (
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"

	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTasksLogic {
	return &ListTasksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListTasksLogic) ListTasks(in *pb.ListFilesTasksRequest) (*pb.TasksResponse, error) {
	items, total, err := l.svcCtx.Management.ListTasks(l.ctx, identity.RawToken(l.ctx), management.FileTaskQuery{Page: in.Page, PageSize: in.PageSize, Keyword: in.Keyword, Status: in.Status})
	if err != nil {
		return nil, err
	}
	result := &pb.TasksResponse{Items: []*pb.TaskRecord{}, Total: total}
	for _, item := range items {
		result.Items = append(result.Items, taskRecord(item))
	}
	return result, nil
}
