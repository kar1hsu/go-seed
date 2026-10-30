package logic

import (
	"github.com/kar1hsu/go-seed/pkg/identity"

	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateFileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateFileLogic {
	return &CreateFileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateFileLogic) CreateFile(in *pb.CreateFileRequest) (*pb.FileResponse, error) {
	item, err := l.svcCtx.Management.CreateFile(l.ctx, identity.RawToken(l.ctx), in.Name, in.ContentType, in.Size)
	if err != nil {
		return nil, err
	}
	return &pb.FileResponse{Record: fileRecord(*item)}, nil
}
