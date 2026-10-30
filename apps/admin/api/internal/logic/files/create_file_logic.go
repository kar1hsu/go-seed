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

type CreateFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFileLogic {
	return &CreateFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateFileLogic) CreateFile(req *types.CreateFileRequest) (resp *types.FileResponse, err error) {
	reply, err := l.svcCtx.System.CreateFile(l.ctx, &pb.CreateFileRequest{Name: req.Name, Size: req.Size, ContentType: req.ContentType})
	if err != nil {
		return nil, err
	}
	result := types.FileData{Record: fileRecord(reply.Record)}
	return &types.FileResponse{Message: "ok", Data: result, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
