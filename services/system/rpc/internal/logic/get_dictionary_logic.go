package logic

import (
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"

	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDictionaryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDictionaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDictionaryLogic {
	return &GetDictionaryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDictionaryLogic) GetDictionary(in *pb.GetDictionaryRequest) (*pb.DictionaryResponse, error) {
	id, err := management.ID(in.Id)
	if err != nil {
		return nil, err
	}
	item, err := l.svcCtx.Management.GetDictionary(l.ctx, identity.RawToken(l.ctx), id)
	if err != nil {
		return nil, err
	}
	return dictionaryResponse(item), nil
}
