package logic

import (
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"

	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFilesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFilesLogic {
	return &ListFilesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListFilesLogic) ListFiles(in *pb.ListFilesTasksRequest) (*pb.FilesResponse, error) {
	items, total, err := l.svcCtx.Management.ListFiles(l.ctx, identity.RawToken(l.ctx), management.FileTaskQuery{Page: in.Page, PageSize: in.PageSize, Keyword: in.Keyword, Status: in.Status})
	if err != nil {
		return nil, err
	}
	result := &pb.FilesResponse{Items: []*pb.FileRecord{}, Total: total}
	for _, item := range items {
		result.Items = append(result.Items, fileRecord(item))
	}
	return result, nil
}
