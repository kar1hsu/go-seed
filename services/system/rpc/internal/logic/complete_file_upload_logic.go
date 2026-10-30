package logic

import (
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"

	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CompleteFileUploadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCompleteFileUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompleteFileUploadLogic {
	return &CompleteFileUploadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CompleteFileUploadLogic) CompleteFileUpload(in *pb.FileTaskIDRequest) (*pb.FileResponse, error) {
	id, err := management.ID(in.Id)
	if err != nil {
		return nil, err
	}
	item, err := l.svcCtx.Management.CompleteFileUpload(l.ctx, identity.RawToken(l.ctx), id)
	if err != nil {
		return nil, err
	}
	return &pb.FileResponse{Record: fileRecord(*item)}, nil
}
