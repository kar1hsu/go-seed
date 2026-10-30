package logic

import (
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"

	"context"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDictionaryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDictionaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDictionaryLogic {
	return &UpdateDictionaryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateDictionaryLogic) UpdateDictionary(in *pb.UpdateDictionaryRequest) (*pb.AdminActionResponse, error) {
	id, err := management.ID(in.Id)
	if err != nil {
		return nil, err
	}
	items, err := dictionaryItems(in.Items)
	if err != nil {
		return nil, err
	}
	input := management.DictionaryInput{Name: in.Name, Description: in.Description, Enabled: in.Enabled, Items: items}
	err = l.svcCtx.Management.UpdateDictionary(l.ctx, identity.RawToken(l.ctx), id, input, in.Revision)
	return &pb.AdminActionResponse{}, err
}
