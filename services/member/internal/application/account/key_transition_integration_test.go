//go:build integration

package account_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/member/internal/application/account"
	"github.com/kar1hsu/go-seed/services/member/internal/repository"
	testidentity "github.com/kar1hsu/go-seed/tests/fixtures/identity"
	"google.golang.org/grpc/codes"
)

func TestSigningKeyTransition(t *testing.T) {
	for _, tc := range []struct {
		name          string
		consumed      bool
		changeRefresh bool
	}{
		{"legacy-active-session", false, true},
		{"jwt-rotation-recovers-lost-response", true, false},
		{"refresh-key-rotation-requires-login-for-lost-response", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := setup(t)
			register(t, f)
			first := login(t, f.a)
			oldConfig := testidentity.Config(4, "member", "test-member", "test-app", 15*time.Minute)
			oldSigner, err := identity.NewSigner(oldConfig)
			if err != nil {
				t.Fatal(err)
			}
			claims, err := oldSigner.Parse(first.AccessToken)
			if err != nil {
				t.Fatal(err)
			}
			// The persisted session and opaque refresh format are independent of
			// the access-token algorithm. Recreate the legacy HS256 access token.
			legacyAccess, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-only-legacy-hmac-secret"))
			if err != nil {
				t.Fatal(err)
			}
			retryKey, err := identity.Random()
			if err != nil {
				t.Fatal(err)
			}
			ctx := identity.WithRefreshKey(context.Background(), retryKey)
			expectedRefresh := ""
			if tc.consumed {
				next, err := f.a.Refresh(ctx, first.RefreshToken)
				if err != nil {
					t.Fatal(err)
				}
				expectedRefresh = next.RefreshToken
			}
			nextConfig := testidentity.Config(25, "member", "test-member", "test-app", 15*time.Minute)
			if !tc.changeRefresh {
				nextConfig.RefreshKey = oldConfig.RefreshKey
			}
			nextSigner, err := identity.NewSigner(nextConfig)
			if err != nil {
				t.Fatal(err)
			}
			upgraded, err := account.New(repository.NewAuth(f.db), nextSigner, 24*time.Hour, func(context.Context, string) error { return nil })
			if err != nil {
				t.Fatal(err)
			}
			_, err = upgraded.Current(context.Background(), legacyAccess)
			code(t, err, codes.Unauthenticated)
			_, err = upgraded.Current(context.Background(), first.AccessToken)
			code(t, err, codes.Unauthenticated)
			next, err := upgraded.Refresh(ctx, first.RefreshToken)
			if tc.consumed && tc.changeRefresh {
				code(t, err, codes.Unauthenticated)
				_, err = f.a.Current(context.Background(), first.AccessToken)
				code(t, err, codes.Unauthenticated)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			nextClaims, err := nextSigner.Parse(next.AccessToken)
			if err != nil {
				t.Fatal(err)
			}
			if nextClaims.Subject != claims.Subject || nextClaims.SessionID != claims.SessionID || nextClaims.Version != claims.Version {
				t.Fatal("key transition changed the session identity")
			}
			if expectedRefresh != "" && next.RefreshToken != expectedRefresh {
				t.Fatal("JWT rotation changed the recovered refresh token")
			}
			if _, err = upgraded.Current(context.Background(), next.AccessToken); err != nil {
				t.Fatal(err)
			}
			recovered, err := upgraded.Refresh(ctx, first.RefreshToken)
			if err != nil {
				t.Fatal(err)
			}
			if recovered.RefreshToken != next.RefreshToken {
				t.Fatal("retry failed to recover the same successor")
			}
			wrongKey, err := identity.Random()
			if err != nil {
				t.Fatal(err)
			}
			_, err = upgraded.Refresh(identity.WithRefreshKey(context.Background(), wrongKey), first.RefreshToken)
			code(t, err, codes.Unauthenticated)
			_, err = upgraded.Current(context.Background(), next.AccessToken)
			code(t, err, codes.Unauthenticated)
		})
	}
}
