package logic

import (
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"

	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthorizeFileUploadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuthorizeFileUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthorizeFileUploadLogic {
	return &AuthorizeFileUploadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AuthorizeFileUploadLogic) AuthorizeFileUpload(in *pb.FileTaskIDRequest) (*pb.FileTransferResponse, error) {
	id, err := management.ID(in.Id)
	if err != nil {
		return nil, err
	}
	item, err := l.svcCtx.Management.AuthorizeFileUpload(l.ctx, identity.RawToken(l.ctx), id)
	if err != nil {
		return nil, err
	}
	return &pb.FileTransferResponse{Key: item.ObjectKey, Record: fileRecord(*item)}, nil
}
