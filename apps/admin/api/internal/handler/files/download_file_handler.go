// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package files

import (
	"context"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/pkg/xhttp"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/logic/files"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Private attachment stream; application/octet-stream, no JSON envelope on success.
func DownloadFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		r = r.WithContext(ctx)
		var req types.FileTaskIDRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.New(codes.InvalidArgument))
			return
		}

		l := files.NewDownloadFileLogic(r.Context(), svcCtx)
		body, record, err := l.DownloadFile(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			defer body.Close()
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": record.Name}))
			w.Header().Set("Content-Length", strconv.FormatInt(record.Size, 10))
			w.Header().Set("Cache-Control", "private, no-store")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			deadline, _ := ctx.Deadline()
			if err := xhttp.Stream(ctx, w, body, deadline); err != nil {
				logx.WithContext(ctx).Errorf("download stream interrupted (%T)", err)
			}
		}
	}
}
