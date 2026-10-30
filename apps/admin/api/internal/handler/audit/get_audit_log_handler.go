// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package audit

import (
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"google.golang.org/grpc/codes"
	"net/http"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/logic/audit"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAuditLogHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetAuditLogRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.New(codes.InvalidArgument))
			return
		}

		l := audit.NewGetAuditLogLogic(r.Context(), svcCtx)
		resp, err := l.GetAuditLog(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
