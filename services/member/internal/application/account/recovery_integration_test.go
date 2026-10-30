//go:build integration

package account_test

import (
	"context"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/member/internal/application/account"
	"github.com/kar1hsu/go-seed/services/member/internal/repository"
	testidentity "github.com/kar1hsu/go-seed/tests/fixtures/identity"
	"google.golang.org/grpc/codes"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRefreshRecoveryConcurrentAndReplay(t *testing.T) {
	f := setup(t)
	register(t, f)
	first := login(t, f.a)
	key, _ := identity.Random()
	ctx := identity.WithRefreshKey(context.Background(), key)
	var wg sync.WaitGroup
	tokens := make(chan string, 2)
	for _, s := range []interface {
		Refresh(context.Context, string) (*account.Session, error)
	}{f.a, f.b} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			next, err := s.Refresh(ctx, first.RefreshToken)
			if err != nil {
				t.Error(err)
				return
			}
			tokens <- next.RefreshToken
		}()
	}
	wg.Wait()
	close(tokens)
	var same string
	for v := range tokens {
		if same != "" && same != v {
			t.Error("concurrent recovery created different successors")
		}
		same = v
	}
	if same == "" {
		t.Fatal("no successor")
	}
	wrong, _ := identity.Random()
	_, err := f.a.Refresh(identity.WithRefreshKey(context.Background(), wrong), first.RefreshToken)
	code(t, err, codes.Unauthenticated)
	_, err = f.a.Refresh(context.Background(), same)
	code(t, err, codes.Unauthenticated)
}

func TestMemberPasswordChangeUsesAccountQuota(t *testing.T) {
	f := setup(t)
	register(t, f)
	signed := login(t, f.a)
	tokens, err := identity.NewSigner(testidentity.Config(4, "member", "test-member", "test-app", 15*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	var quotaKey string
	limited, err := account.New(repository.NewAuth(f.db), tokens, 24*time.Hour, func(_ context.Context, key string) error { quotaKey = key; return xerr.New(codes.ResourceExhausted) })
	if err != nil {
		t.Fatal(err)
	}
	err = limited.ChangePassword(context.Background(), signed.AccessToken, "incorrect", "replacement-password")
	code(t, err, codes.ResourceExhausted)
	if !strings.HasPrefix(quotaKey, "password:") {
		t.Fatal("password verification did not use authenticated account quota")
	}
}
