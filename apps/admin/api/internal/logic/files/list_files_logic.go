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

type ListFilesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFilesLogic {
	return &ListFilesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListFilesLogic) ListFiles(req *types.ListFilesTasksRequest) (resp *types.FilesResponse, err error) {
	reply, err := l.svcCtx.System.ListFiles(l.ctx, &pb.ListFilesTasksRequest{Page: req.Page, PageSize: req.PageSize, Keyword: req.Keyword, Status: req.Status})
	if err != nil {
		return nil, err
	}
	result := types.FilesData{Items: []types.FileRecord{}, Total: reply.Total, Page: req.Page, PageSize: req.PageSize}
	for _, item := range reply.Items {
		result.Items = append(result.Items, fileRecord(item))
	}
	return &types.FilesResponse{Message: "ok", Data: result, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
