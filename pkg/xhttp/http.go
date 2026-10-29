package xhttp

import (
	"context"
	"net/http"
	"time"

	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/trace"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	TraceID string `json:"traceId"`
}

func ErrorHandler(ctx context.Context, err error) (int, any) {
	public := xerr.Describe(err)
	if public.HTTPStatus >= 500 {
		logx.WithContext(ctx).Errorf("request failed: %v", err)
	}
	return public.HTTPStatus, ErrorResponse{Code: public.Code, Message: public.Message, TraceID: trace.TraceIDFromContext(ctx)}
}

// RegisterHealth keeps process liveness independent from downstream readiness.
func RegisterHealth(server *rest.Server, ready func(context.Context) error) {
	server.AddRoutes([]rest.Route{
		{Method: http.MethodGet, Path: "/healthz", Handler: func(w http.ResponseWriter, r *http.Request) {
			httpx.OkJsonCtx(r.Context(), w, map[string]string{"status": "ok"})
		}},
		{Method: http.MethodGet, Path: "/readyz", Handler: func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := ready(ctx); err != nil {
				httpx.WriteJsonCtx(ctx, w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
				return
			}
			httpx.OkJsonCtx(ctx, w, map[string]string{"status": "ok"})
		}},
	})
}
