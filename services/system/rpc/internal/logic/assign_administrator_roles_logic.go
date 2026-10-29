package logic

import (
	"context"

	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
)

type AssignAdministratorRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignAdministratorRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignAdministratorRolesLogic {
	return &AssignAdministratorRolesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AssignAdministratorRolesLogic) AssignAdministratorRoles(in *pb.AssignAdministratorRolesRequest) (*pb.AdminActionResponse, error) {
	id, err := management.ID(in.Id)
	if err != nil {
		return nil, err
	}
	ids := []uint64{}
	for _, raw := range in.RoleIds {
		roleID, err := management.ID(raw)
		if err != nil {
			return nil, err
		}
		ids = append(ids, roleID)
	}
	err = l.svcCtx.Management.AssignRoles(l.ctx, identity.RawToken(l.ctx), id, ids, in.Revision)
	return &pb.AdminActionResponse{}, err
}
