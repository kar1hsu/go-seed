// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package management

import (
	"context"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/trace"
)

type CreateAdministratorLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateAdministratorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAdministratorLogic {
	return &CreateAdministratorLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateAdministratorLogic) CreateAdministrator(req *types.CreateAdministratorRequest) (resp *types.AdministratorResponse, err error) {
	reply, err := l.svcCtx.System.CreateAdministrator(l.ctx, &pb.CreateAdministratorRequest{Username: req.Username, DisplayName: req.DisplayName, Password: req.Password})
	if err != nil {
		return nil, err
	}
	return &types.AdministratorResponse{Message: "ok", Data: administratorRecord(reply.Admin), TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
