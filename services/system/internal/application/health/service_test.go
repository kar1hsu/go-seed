package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestReadyChecksDependenciesAndPreservesCancellation(t *testing.T) {
	ok := func(context.Context) error { return nil }
	failure := errors.New("offline")
	for _, dependency := range []string{"mysql", "redis"} {
		t.Run(dependency, func(t *testing.T) {
			bad := func(context.Context) error { return failure }
			mysql, redis := PingFunc(ok), PingFunc(ok)
			if dependency == "mysql" {
				mysql = bad
			} else {
				redis = bad
			}
			if err := New(mysql, redis).Ready(context.Background()); !errors.Is(err, failure) {
				t.Fatalf("unexpected: %v", err)
			}
		})
	}
	if err := New(ok, ok).Ready(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	wait := func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }
	if err := New(wait, ok).Ready(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
	bounded := func(ctx context.Context) error {
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > 1500*time.Millisecond {
			t.Fatal("readiness deadline missing")
		}
		return nil
	}
	if err := New(bounded, bounded).Ready(context.Background()); err != nil {
		t.Fatal(err)
	}
}
