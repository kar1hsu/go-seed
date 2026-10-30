package identity

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

type refreshKey struct{}

func ValidRefreshKey(value string) bool {
	if value == "" {
		return true
	}
	b, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && len(b) == 32 && base64.RawURLEncoding.EncodeToString(b) == value
}
func WithRefreshKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, refreshKey{}, key)
}
func RefreshKey(ctx context.Context) string { v, _ := ctx.Value(refreshKey{}).(string); return v }
func RefreshKeyDigest(key string) string {
	if key == "" {
		return ""
	}
	v := sha256.Sum256([]byte(key))
	return hex.EncodeToString(v[:])
}

// RefreshSecret derives an idempotent successor without storing bearer credentials.
// The purpose and identity kind separate this PRF from JWT signing and other domains.
func (t *Signer) RefreshSecret(previous, key string) string {
	m := hmac.New(sha256.New, t.refreshKey)
	m.Write([]byte("goseed/refresh/v1\x00" + t.config.Kind + "\x00" + previous + "\x00" + key))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}
