// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package files

import (
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"google.golang.org/grpc/codes"
	"io"

	"context"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Private attachment stream; application/octet-stream, no JSON envelope on success.
func NewDownloadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadFileLogic {
	return &DownloadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadFileLogic) DownloadFile(req *types.FileTaskIDRequest) (io.ReadCloser, *pb.FileRecord, error) {
	ref, err := l.svcCtx.System.DownloadFile(l.ctx, &pb.FileTaskIDRequest{Id: req.Id})
	if err != nil {
		return nil, nil, err
	}
	r, err := l.svcCtx.Objects.Open(l.ctx, ref.Key)
	if err != nil {
		return nil, nil, xerr.New(codes.Unavailable)
	}
	return r, ref.Record, nil
}
