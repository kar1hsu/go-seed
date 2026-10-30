package middleware

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kar1hsu/go-seed/pkg/storage"
)

// FinishUploadBody runs outside go-zero's buffered timeout writer so it can set
// a real connection deadline. Draining a rejected upload prevents a TCP reset
// from replacing its 401/403 response with a reverse-proxy 502.
func FinishUploadBody(next http.Handler) http.Handler {
	return finishUploadBody(next, 5*time.Second)
}

func finishUploadBody(next http.Handler, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/admin-api/v1/system/files/") && strings.HasSuffix(r.URL.Path, "/content") {
			defer func() {
				control := http.NewResponseController(w)
				if err := control.SetReadDeadline(time.Now().Add(timeout)); err != nil {
					return
				}
				defer control.SetReadDeadline(time.Time{})
				// Rejects are still bounded by the upload limit; never buffer bytes.
				_, _ = io.Copy(io.Discard, io.LimitReader(r.Body, storage.MaxSize+1))
			}()
		}
		next.ServeHTTP(w, r)
	})
}
