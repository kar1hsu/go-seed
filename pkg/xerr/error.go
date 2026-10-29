// Package xerr defines safe public errors at transport boundaries.
package xerr

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Public struct {
	HTTPStatus int
	Code       int
	Message    string
}

// Describe never exposes arbitrary database, network or RPC error text.
func Describe(err error) Public {
	code := status.Code(err)
	if errors.Is(err, context.DeadlineExceeded) {
		code = codes.DeadlineExceeded
	}
	if errors.Is(err, context.Canceled) {
		code = codes.Canceled
	}
	switch code {
	case codes.InvalidArgument:
		return Public{400, 100400, "invalid request"}
	case codes.Unauthenticated:
		return Public{401, 100401, "authentication required"}
	case codes.PermissionDenied:
		return Public{403, 100403, "permission denied"}
	case codes.NotFound:
		return Public{404, 100404, "resource not found"}
	case codes.AlreadyExists, codes.FailedPrecondition, codes.Aborted:
		return Public{409, 100409, "resource conflict"}
	case codes.ResourceExhausted:
		return Public{429, 100429, "too many requests"}
	case codes.Unavailable:
		return Public{503, 100503, "service unavailable"}
	case codes.DeadlineExceeded:
		return Public{504, 100504, "request timed out"}
	case codes.Canceled:
		return Public{408, 100408, "request canceled"}
	default:
		return Public{http.StatusInternalServerError, 100500, "internal server error"}
	}
}

func New(code codes.Code) error {
	public := Describe(status.Error(code, ""))
	return status.Error(code, public.Message)
}
