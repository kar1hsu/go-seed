package logic

import (
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"

	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateDictionaryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateDictionaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDictionaryLogic {
	return &CreateDictionaryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateDictionaryLogic) CreateDictionary(in *pb.CreateDictionaryRequest) (*pb.DictionaryResponse, error) {
	items, err := dictionaryItems(in.Items)
	if err != nil {
		return nil, err
	}
	input := management.DictionaryInput{Name: in.Name, Description: in.Description, Enabled: in.Enabled, Items: items}
	item, err := l.svcCtx.Management.CreateDictionary(l.ctx, identity.RawToken(l.ctx), in.Code, input)
	if err != nil {
		return nil, err
	}
	return dictionaryResponse(item), nil
}
