// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package system

import (
	"net/http"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/logic/system"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Probe API to system-rpc connectivity
func PingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := system.NewPingLogic(r.Context(), svcCtx)
		resp, err := l.Ping()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
