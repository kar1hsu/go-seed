// Package identity provides deterministic test-only signing material. Never use
// these values in service configuration or production.
package identity

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"time"

	"github.com/kar1hsu/go-seed/pkg/identity"
)

func Config(marker byte, kind, issuer, audience string, ttl time.Duration) identity.SignerConfig {
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{marker}, ed25519.SeedSize))
	return identity.SignerConfig{
		Config:     identity.Config{Kind: kind, PublicKey: base64.RawURLEncoding.EncodeToString(private.Public().(ed25519.PublicKey)), Issuer: issuer, Audience: audience},
		PrivateKey: base64.RawURLEncoding.EncodeToString(private),
		RefreshKey: base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{marker ^ 0xff}, 32)),
		AccessTTL:  ttl,
	}
}
