// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package management

import (
	"net/http"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/logic/management"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListPermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := management.NewListPermissionsLogic(r.Context(), svcCtx)
		resp, err := l.ListPermissions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
