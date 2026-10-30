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

type GetDictionaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDictionaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDictionaryLogic {
	return &GetDictionaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDictionaryLogic) GetDictionary(req *types.GetDictionaryRequest) (resp *types.DictionaryResponse, err error) {
	reply, err := l.svcCtx.System.GetDictionary(l.ctx, &pb.GetDictionaryRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}
	return &types.DictionaryResponse{Message: "ok", Data: dictionaryDetail(reply), TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
