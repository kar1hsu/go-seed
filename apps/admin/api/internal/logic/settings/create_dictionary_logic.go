// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package settings

import (
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/trace"

	"context"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateDictionaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateDictionaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDictionaryLogic {
	return &CreateDictionaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDictionaryLogic) CreateDictionary(req *types.CreateDictionaryRequest) (resp *types.DictionaryResponse, err error) {
	reply, err := l.svcCtx.System.CreateDictionary(l.ctx, &pb.CreateDictionaryRequest{Name: req.Name, Description: req.Description, Enabled: req.Enabled, Items: dictionaryItems(req.Items), Code: req.Code})
	if err != nil {
		return nil, err
	}
	return &types.DictionaryResponse{Message: "ok", Data: dictionaryDetail(reply), TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
