//go:build integration

package auth

import (
	"context"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"google.golang.org/grpc/codes"
	"strings"
	"testing"
	"time"
)

func TestRefreshRecoveryKeepsReplayDetection(t *testing.T) {
	f := setup(t)
	key, _ := identity.Random()
	ctx := identity.WithRefreshKey(context.Background(), key)
	first := login(t, f.first, f.username, testPassword)
	rotated, err := f.first.Refresh(ctx, first.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	recovered, err := f.second.Refresh(ctx, first.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.RefreshToken != rotated.RefreshToken {
		t.Fatal("lost response did not recover the same successor")
	}
	var records []model.AdminRefreshToken
	if err = f.db.Find(&records, "session_id = ?", strings.Split(first.RefreshToken, ".")[0]).Error; err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatal("recovery created extra tokens")
	}
	for _, r := range records {
		if r.RetryKey == key {
			t.Fatal("stored raw request key")
		}
	}
	nextKey, _ := identity.Random()
	_, err = f.first.Refresh(identity.WithRefreshKey(context.Background(), nextKey), rotated.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.first.Refresh(ctx, first.RefreshToken)
	wantCode(t, err, codes.Unauthenticated)
	_, err = f.first.Current(context.Background(), recovered.AccessToken)
	wantCode(t, err, codes.Unauthenticated)
}
func TestRefreshRecoveryRejectsWrongKeyAndExpiredWindow(t *testing.T) {
	for _, expired := range []bool{false, true} {
		t.Run(map[bool]string{true: "expired", false: "different-key"}[expired], func(t *testing.T) {
			f := setup(t)
			first := login(t, f.first, f.username, testPassword)
			key, _ := identity.Random()
			ctx := identity.WithRefreshKey(context.Background(), key)
			next, err := f.first.Refresh(ctx, first.RefreshToken)
			if err != nil {
				t.Fatal(err)
			}
			if expired {
				if err = f.db.Model(&model.AdminRefreshToken{}).Where("digest = ?", digest(first.RefreshToken)).Update("consumed_at", time.Now().Add(-2*time.Minute)).Error; err != nil {
					t.Fatal(err)
				}
			} else {
				key, _ = identity.Random()
				ctx = identity.WithRefreshKey(context.Background(), key)
			}
			_, err = f.second.Refresh(ctx, first.RefreshToken)
			wantCode(t, err, codes.Unauthenticated)
			_, err = f.first.Current(context.Background(), next.AccessToken)
			wantCode(t, err, codes.Unauthenticated)
		})
	}
}
func TestPasswordChangeIsLimitedBeforeBcrypt(t *testing.T) {
	f := setup(t)
	signed := login(t, f.first, f.username, testPassword)
	var key string
	f.first.allowLogin = func(_ context.Context, k string) error { key = k; return xerr.New(codes.ResourceExhausted) }
	err := f.first.ChangePassword(context.Background(), signed.AccessToken, "wrong", "new-password-that-is-valid")
	wantCode(t, err, codes.ResourceExhausted)
	if !strings.HasPrefix(key, "password:") {
		t.Fatal("wrong password quota key")
	}
}
