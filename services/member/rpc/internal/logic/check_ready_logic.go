package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/member/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
	"google.golang.org/grpc/codes"
)

type CheckReadyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckReadyLogic(ctx context.Context, s *svc.ServiceContext) *CheckReadyLogic {
	return &CheckReadyLogic{ctx, s}
}
func (l *CheckReadyLogic) CheckReady(in *pb.Empty) (*pb.ReadyResponse, error) {
	if err := l.svcCtx.Ready(l.ctx); err != nil {
		return nil, xerr.New(codes.Unavailable)
	}
	return &pb.ReadyResponse{Ready: true}, nil
}
