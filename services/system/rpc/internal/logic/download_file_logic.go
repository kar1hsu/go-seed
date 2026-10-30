package logic

import (
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"

	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadFileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDownloadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadFileLogic {
	return &DownloadFileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DownloadFileLogic) DownloadFile(in *pb.FileTaskIDRequest) (*pb.FileTransferResponse, error) {
	id, err := management.ID(in.Id)
	if err != nil {
		return nil, err
	}
	item, err := l.svcCtx.Management.DownloadFile(l.ctx, identity.RawToken(l.ctx), id)
	if err != nil {
		return nil, err
	}
	return &pb.FileTransferResponse{Key: item.ObjectKey, Record: fileRecord(*item)}, nil
}
