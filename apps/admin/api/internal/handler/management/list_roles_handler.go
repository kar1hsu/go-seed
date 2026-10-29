// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package management

import (
	"net/http"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/logic/management"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
)

func ListRolesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListManagementRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.New(codes.InvalidArgument))
			return
		}

		l := management.NewListRolesLogic(r.Context(), svcCtx)
		resp, err := l.ListRoles(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
