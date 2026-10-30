package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/rpcx"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"google.golang.org/grpc/codes"
	"strconv"
	"strings"
)

type CheckAccessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckAccessLogic(ctx context.Context, s *svc.ServiceContext) *CheckAccessLogic {
	return &CheckAccessLogic{ctx, s}
}
func (l *CheckAccessLogic) CheckAccess(in *pb.CheckAccessRequest) (*pb.CheckAccessResponse, error) {
	allowed := false
	for _, c := range l.svcCtx.Config.DomainCallers {
		if c.Name == rpcx.ServiceName(l.ctx) && strings.HasPrefix(in.Permission, c.PermissionPrefix) {
			allowed = true
		}
	}
	if !allowed {
		return nil, xerr.New(codes.PermissionDenied)
	}
	a, err := l.svcCtx.Management.Current(l.ctx, identity.RawToken(l.ctx))
	if err != nil {
		return nil, err
	}
	if err = a.Require(in.Permission); err != nil {
		return nil, err
	}
	return &pb.CheckAccessResponse{AdminId: strconv.FormatUint(a.Admin.ID, 10)}, nil
}
