package identity_test

import (
	"crypto/ed25519"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/kar1hsu/go-seed/pkg/identity"
	testidentity "github.com/kar1hsu/go-seed/tests/fixtures/identity"
)

func TestAccessTokenClaims(t *testing.T) {
	cfg := testidentity.Config(5, "admin", "go-seed-admin", "go-seed-admin-api", 15*time.Minute)
	signer, err := identity.NewSigner(cfg)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := identity.NewVerifier(cfg.Config)
	if err != nil {
		t.Fatal(err)
	}
	sid, _ := identity.Random()
	raw, _, err := signer.Issue(42, sid, 3)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := verifier.Parse(raw)
	if err != nil || claims.Subject != "42" {
		t.Fatal("valid token rejected", err)
	}
	private, _ := base64.RawURLEncoding.DecodeString(cfg.PrivateKey)
	cases := map[string]func(*identity.Claims){
		"issuer":         func(c *identity.Claims) { c.Issuer = "member" },
		"audience":       func(c *identity.Claims) { c.Audience = jwt.ClaimStrings{"app"} },
		"subject type":   func(c *identity.Claims) { c.Kind = "member" },
		"subject ID":     func(c *identity.Claims) { c.Subject = "0" },
		"session":        func(c *identity.Claims) { c.SessionID = "" },
		"version":        func(c *identity.Claims) { c.Version = 0 },
		"missing jti":    func(c *identity.Claims) { c.ID = "" },
		"missing expiry": func(c *identity.Claims) { c.ExpiresAt = nil },
		"missing issue":  func(c *identity.Claims) { c.IssuedAt = nil },
		"missing nbf":    func(c *identity.Claims) { c.NotBefore = nil },
		"expired":        func(c *identity.Claims) { c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute)) },
		"future issue":   func(c *identity.Claims) { c.IssuedAt = jwt.NewNumericDate(time.Now().Add(time.Hour)) },
		"future nbf":     func(c *identity.Claims) { c.NotBefore = jwt.NewNumericDate(time.Now().Add(time.Hour)) },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			current, err := verifier.Parse(raw)
			if err != nil {
				t.Fatal(err)
			}
			change(current)
			invalid, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, current).SignedString(ed25519.PrivateKey(private))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := verifier.Parse(invalid); err == nil {
				t.Fatal("invalid claims accepted")
			}
		})
	}
	public, _ := base64.RawURLEncoding.DecodeString(cfg.PublicKey)
	for _, method := range []jwt.SigningMethod{jwt.SigningMethodHS256, jwt.SigningMethodHS384, jwt.SigningMethodNone} {
		key := any(public) // A public key must never be accepted as an HMAC secret.
		if method == jwt.SigningMethodNone {
			key = jwt.UnsafeAllowNoneSignatureType
		}
		invalid, err := jwt.NewWithClaims(method, claims).SignedString(key)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := verifier.Parse(invalid); err == nil {
			t.Fatal("wrong algorithm accepted", method.Alg())
		}
	}
	wrong := testidentity.Config(6, "admin", cfg.Issuer, cfg.Audience, cfg.AccessTTL)
	wrongPrivate, _ := base64.RawURLEncoding.DecodeString(wrong.PrivateKey)
	invalid, _ := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(ed25519.PrivateKey(wrongPrivate))
	if _, err := verifier.Parse(invalid); err == nil {
		t.Fatal("wrong key accepted")
	}
	for _, value := range []string{"", strings.Repeat("a", 4097), "invalid"} {
		if _, err := verifier.Parse(value); err == nil {
			t.Fatal("malformed token accepted")
		}
	}
	// API verifiers have neither signing nor refresh-secret derivation capability.
	if _, ok := any(verifier).(interface {
		Issue(uint64, string, uint64) (string, time.Time, error)
	}); ok {
		t.Fatal("verifier can issue tokens")
	}
	if _, ok := any(verifier).(interface{ RefreshSecret(string, string) string }); ok {
		t.Fatal("verifier can derive refresh credentials")
	}
}

func TestMemberAndAdminKindsAreIsolated(t *testing.T) {
	// Even accidental reuse of a key/issuer/audience cannot bridge identity kinds.
	c := testidentity.Config(7, "admin", "shared-test", "shared-test", 15*time.Minute)
	admin, err := identity.NewSigner(c)
	if err != nil {
		t.Fatal(err)
	}
	c.Kind = "member"
	member, err := identity.NewSigner(c)
	if err != nil {
		t.Fatal(err)
	}
	sid, _ := identity.Random()
	a, _, _ := admin.Issue(1, sid, 1)
	m, _, _ := member.Issue(1, sid, 1)
	if _, err = member.Parse(a); err == nil {
		t.Fatal("member accepted admin token")
	}
	if _, err = admin.Parse(m); err == nil {
		t.Fatal("admin accepted member token")
	}
	if _, err = member.Parse(m); err != nil {
		t.Fatal(err)
	}
	c.Kind = "unknown"
	if _, err = identity.NewSigner(c); err == nil {
		t.Fatal("unknown kind accepted")
	}
}

func TestSigningConfigurationFailsClosed(t *testing.T) {
	cfg := testidentity.Config(8, "admin", "issuer", "audience", time.Minute)
	other := testidentity.Config(9, "admin", "issuer", "audience", time.Minute)
	for name, change := range map[string]func(*identity.SignerConfig){
		"missing public":         func(c *identity.SignerConfig) { c.PublicKey = "" },
		"malformed public":       func(c *identity.SignerConfig) { c.PublicKey = "invalid" },
		"missing private":        func(c *identity.SignerConfig) { c.PrivateKey = "" },
		"public used as private": func(c *identity.SignerConfig) { c.PrivateKey = c.PublicKey },
		"mismatched pair":        func(c *identity.SignerConfig) { c.PublicKey = other.PublicKey },
		"corrupt private": func(c *identity.SignerConfig) {
			b, _ := base64.RawURLEncoding.DecodeString(c.PrivateKey)
			b[40] ^= 1
			c.PrivateKey = base64.RawURLEncoding.EncodeToString(b)
		},
		"missing refresh":   func(c *identity.SignerConfig) { c.RefreshKey = "" },
		"public as refresh": func(c *identity.SignerConfig) { c.RefreshKey = c.PublicKey },
		"seed as refresh": func(c *identity.SignerConfig) {
			b, _ := base64.RawURLEncoding.DecodeString(c.PrivateKey)
			c.RefreshKey = base64.RawURLEncoding.EncodeToString(b[:32])
		},
		"bad ttl": func(c *identity.SignerConfig) { c.AccessTTL = 2 * time.Hour },
	} {
		t.Run(name, func(t *testing.T) {
			c := cfg
			change(&c)
			if _, err := identity.NewSigner(c); err == nil {
				t.Fatal("invalid config accepted")
			}
		})
	}
}

func TestRefreshRecoveryKeyIsIndependentOfJWTKey(t *testing.T) {
	c := testidentity.Config(10, "admin", "issuer", "audience", time.Minute)
	original, err := identity.NewSigner(c)
	if err != nil {
		t.Fatal(err)
	}
	rotated := testidentity.Config(11, "admin", "issuer", "audience", time.Minute)
	rotated.RefreshKey = c.RefreshKey
	next, err := identity.NewSigner(rotated)
	if err != nil {
		t.Fatal(err)
	}
	want := original.RefreshSecret("previous-cookie", "retry-key")
	if next.RefreshSecret("previous-cookie", "retry-key") != want {
		t.Fatal("JWT key rotation broke refresh recovery")
	}
	rotated.RefreshKey = testidentity.Config(12, "admin", "issuer", "audience", time.Minute).RefreshKey
	replaced, _ := identity.NewSigner(rotated)
	if replaced.RefreshSecret("previous-cookie", "retry-key") == want {
		t.Fatal("refresh key was ignored")
	}
	c.Kind = "member"
	member, _ := identity.NewSigner(c)
	if member.RefreshSecret("previous-cookie", "retry-key") == want {
		t.Fatal("refresh identity domains overlap")
	}
	if original.RefreshSecret("other-cookie", "retry-key") == want || original.RefreshSecret("previous-cookie", "other-retry") == want {
		t.Fatal("refresh recovery inputs were ignored")
	}
}
