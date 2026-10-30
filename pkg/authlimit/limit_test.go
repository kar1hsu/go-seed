package authlimit

import (
	"context"
	"github.com/kar1hsu/go-seed/pkg/clientip"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func quota(n int) Take {
	counts := map[string]int{}
	return func(_ context.Context, k string) error {
		counts[k]++
		if counts[k] > n {
			return xerr.New(codes.ResourceExhausted)
		}
		return nil
	}
}
func TestSourceQuotasAndPasswordAccount(t *testing.T) {
	l := &Limiter{quota(3), quota(2), quota(1), quota(1)}
	a := clientip.With(context.Background(), "192.0.2.1")
	b := clientip.With(context.Background(), "192.0.2.2")
	for _, k := range []string{"admin", "random1", "random2"} {
		if err := l.Allow(a, k); err != nil {
			t.Fatal(err)
		}
	}
	if status.Code(l.Allow(a, "random3")) != codes.ResourceExhausted {
		t.Fatal("IP quota bypassed with random accounts")
	}
	if err := l.Allow(b, "admin"); err != nil {
		t.Fatal("one source locked another out", err)
	}
	if err := l.Allow(a, "register:a"); err != nil {
		t.Fatal(err)
	}
	if status.Code(l.Allow(a, "register:b")) != codes.ResourceExhausted {
		t.Fatal("registration quota bypassed")
	}
	if err := l.Allow(b, "register:a"); err != nil {
		t.Fatal(err)
	}
	if err := l.Allow(a, "password:42"); err != nil {
		t.Fatal(err)
	}
	if status.Code(l.Allow(b, "password:42")) != codes.ResourceExhausted {
		t.Fatal("authenticated account password quota bypassed")
	}
	if status.Code(l.Allow(context.Background(), "admin")) != codes.Unauthenticated {
		t.Fatal("missing trusted IP accepted")
	}
}
func TestGateBoundsWorkAndCancellation(t *testing.T) {
	g := NewGate(1)
	release, err := g.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.Acquire(context.Background()); status.Code(err) != codes.ResourceExhausted {
		t.Fatal("unbounded password work")
	}
	release()
	release, err = g.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	release()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = g.Acquire(ctx); err != context.Canceled {
		t.Fatal(err)
	}
}
