// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package member

import (
	"net/http"

	"github.com/kar1hsu/go-seed/apps/app/api/internal/logic/member"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Read the current member; no client-supplied user ID
func MemberCurrentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := member.NewMemberCurrentLogic(r.Context(), svcCtx)
		resp, err := l.MemberCurrent()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
