package xerr

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPublicMapping(t *testing.T) {
	tests := []struct {
		code codes.Code
		http int
	}{
		{codes.InvalidArgument, 400}, {codes.Unauthenticated, 401}, {codes.PermissionDenied, 403},
		{codes.NotFound, 404}, {codes.AlreadyExists, 409}, {codes.FailedPrecondition, 409}, {codes.Aborted, 409},
		{codes.ResourceExhausted, 429}, {codes.Unavailable, 503}, {codes.DeadlineExceeded, 504}, {codes.Canceled, 408}, {codes.Internal, 500}, {codes.Unknown, 500},
	}
	for _, tt := range tests {
		t.Run(tt.code.String(), func(t *testing.T) {
			result := Describe(status.Error(tt.code, "password=secret"))
			if result.HTTPStatus != tt.http || result.Code != 100000+tt.http {
				t.Fatalf("unexpected mapping: %+v", result)
			}
			if result.Message == "password=secret" {
				t.Fatal("leaked internal detail")
			}
		})
	}
	for _, tt := range []struct {
		err  error
		code int
	}{
		{errors.New("secret DSN"), 100500}, {fmt.Errorf("wrapped: %w", context.DeadlineExceeded), 100504}, {context.Canceled, 100408},
	} {
		if got := Describe(tt.err); got.Code != tt.code {
			t.Fatalf("%v: %+v", tt.err, got)
		}
	}
}
