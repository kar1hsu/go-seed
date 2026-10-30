package logic

import (
	"github.com/kar1hsu/go-seed/pkg/identity"

	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDictionariesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListDictionariesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDictionariesLogic {
	return &ListDictionariesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListDictionariesLogic) ListDictionaries(in *pb.ListManagementRequest) (*pb.ListDictionariesResponse, error) {
	items, total, err := l.svcCtx.Management.ListDictionaries(l.ctx, identity.RawToken(l.ctx), managementQuery(in))
	if err != nil {
		return nil, err
	}
	result := &pb.ListDictionariesResponse{Items: []*pb.DictionaryRecord{}, Total: total}
	for _, item := range items {
		result.Items = append(result.Items, dictionaryRecord(item))
	}
	return result, nil
}
