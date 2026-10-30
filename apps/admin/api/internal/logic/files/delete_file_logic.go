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

type DeleteFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFileLogic {
	return &DeleteFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFileLogic) DeleteFile(req *types.FileTaskIDRequest) (resp *types.ActionResponse, err error) {
	_, err = l.svcCtx.System.DeleteFile(l.ctx, &pb.FileTaskIDRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}
	return &types.ActionResponse{Message: "ok", TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
