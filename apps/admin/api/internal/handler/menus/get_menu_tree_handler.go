// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package menus

import (
	"net/http"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/logic/menus"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Read the editable navigation tree
func GetMenuTreeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := menus.NewGetMenuTreeLogic(r.Context(), svcCtx)
		resp, err := l.GetMenuTree()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
