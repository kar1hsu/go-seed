package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/services/member/internal/application/management"
	"github.com/kar1hsu/go-seed/services/member/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
	"strconv"
	"time"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserLogic(ctx context.Context, s *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{ctx, s}
}
func (l *GetUserLogic) GetUser(in *pb.UserID) (*pb.UserDetail, error) {
	id, err := management.ID(in.Id)
	if err != nil {
		return nil, err
	}
	u, events, err := l.svcCtx.Management.Get(l.ctx, id)
	if err != nil {
		return nil, err
	}
	result := &pb.UserDetail{User: user(u), RecentEvents: []*pb.AuditRecord{}}
	for _, e := range events {
		result.RecentEvents = append(result.RecentEvents, &pb.AuditRecord{Id: strconv.FormatUint(e.ID, 10), Action: e.Action, ActorKind: e.ActorKind, ActorId: strconv.FormatUint(e.ActorID, 10), CreatedAt: e.CreatedAt.UTC().Format(time.RFC3339), TraceId: e.TraceID})
	}
	return result, nil
}
