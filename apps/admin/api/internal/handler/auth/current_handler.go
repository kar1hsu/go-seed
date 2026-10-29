// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package auth

import (
	"net/http"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/logic/auth"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Read the current administrator; checks live session state
func CurrentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := auth.NewCurrentLogic(r.Context(), svcCtx)
		resp, err := l.Current()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
