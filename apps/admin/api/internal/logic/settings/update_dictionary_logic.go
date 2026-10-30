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

type UpdateDictionaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateDictionaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDictionaryLogic {
	return &UpdateDictionaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDictionaryLogic) UpdateDictionary(req *types.UpdateDictionaryRequest) (resp *types.ActionResponse, err error) {
	_, err = l.svcCtx.System.UpdateDictionary(l.ctx, &pb.UpdateDictionaryRequest{Name: req.Name, Description: req.Description, Enabled: req.Enabled, Items: dictionaryItems(req.Items), Id: req.Id, Revision: req.Revision})
	if err != nil {
		return nil, err
	}
	return &types.ActionResponse{Message: "ok", Data: types.ActionData{Success: true}, TraceId: trace.TraceIDFromContext(l.ctx)}, nil
}
