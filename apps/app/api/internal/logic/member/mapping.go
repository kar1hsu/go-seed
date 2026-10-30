package member

import (
	"context"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/authhttp"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/types"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
	"github.com/zeromicro/go-zero/core/trace"
)

func profileResponse(ctx context.Context, p *pb.Profile) *types.MemberProfileResponse {
	return &types.MemberProfileResponse{Message: "ok", Data: authhttp.Profile(p), TraceId: trace.TraceIDFromContext(ctx)}
}
func action(ctx context.Context) *types.MemberActionResponse {
	return &types.MemberActionResponse{Message: "ok", Data: types.MemberActionData{Success: true}, TraceId: trace.TraceIDFromContext(ctx)}
}
