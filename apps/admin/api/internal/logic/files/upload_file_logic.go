// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package files

import (
	"errors"
	"github.com/kar1hsu/go-seed/pkg/storage"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/trace"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc/codes"
	"io"
	"time"

	"context"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Raw application/octet-stream upload, at most 10 MiB; metadata must be created first.
func NewUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadFileLogic {
	return &UploadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadFileLogic) UploadFile(req *types.FileTaskIDRequest, body io.Reader, length int64) (resp *types.FileResponse, err error) {
	ref, err := l.svcCtx.System.AuthorizeFileUpload(l.ctx, &pb.FileTaskIDRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}
	if length >= 0 && length != ref.Record.Size {
		return nil, xerr.New(codes.InvalidArgument)
	}
	err = l.svcCtx.Objects.Put(l.ctx, ref.Key, body, ref.Record.Size)
	if errors.Is(err, storage.ErrSize) {
		return nil, xerr.New(codes.InvalidArgument)
	}
	if err != nil && !errors.Is(err, storage.ErrExists) {
		return nil, xerr.New(codes.Unavailable)
	}
	reply, err := l.svcCtx.System.CompleteFileUpload(l.ctx, &pb.FileTaskIDRequest{Id: req.Id}, zrpc.WithCallTimeout(30*time.Second))
	if err != nil {
		return nil, err
	}
	return &types.FileResponse{Message: "ok", Data: types.FileData{Record: fileRecord(reply.Record)}, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
