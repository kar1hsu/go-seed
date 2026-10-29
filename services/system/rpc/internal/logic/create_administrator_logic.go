package logic

import (
	"context"

	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAdministratorLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateAdministratorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAdministratorLogic {
	return &CreateAdministratorLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateAdministratorLogic) CreateAdministrator(in *pb.CreateAdministratorRequest) (*pb.AdministratorResponse, error) {
	item, err := l.svcCtx.Management.CreateAdministrator(l.ctx, identity.RawToken(l.ctx), in.Username, in.DisplayName, in.Password)
	if err != nil {
		return nil, err
	}
	return &pb.AdministratorResponse{Admin: administratorRecord(*item)}, nil
}
