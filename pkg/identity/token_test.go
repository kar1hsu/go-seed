package identity

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestAccessTokenClaims(t *testing.T) {
	cfg := Config{Secret: strings.Repeat("s", 48), Issuer: "go-seed-admin", Audience: "go-seed-admin-api", AccessTTL: 15 * time.Minute}
	tokens, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	sid, _ := Random()
	raw, _, err := tokens.Issue(42, sid, 3)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := tokens.Parse(raw)
	if err != nil || claims.Subject != "42" {
		t.Fatal("valid token rejected", err)
	}
	cases := map[string]func(*Claims){
		"issuer":         func(c *Claims) { c.Issuer = "member" },
		"audience":       func(c *Claims) { c.Audience = jwt.ClaimStrings{"app"} },
		"subject type":   func(c *Claims) { c.Kind = "member" },
		"subject ID":     func(c *Claims) { c.Subject = "0" },
		"session":        func(c *Claims) { c.SessionID = "" },
		"version":        func(c *Claims) { c.Version = 0 },
		"missing expiry": func(c *Claims) { c.ExpiresAt = nil },
		"expired":        func(c *Claims) { c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute)) },
		"future issue":   func(c *Claims) { c.IssuedAt = jwt.NewNumericDate(time.Now().Add(time.Hour)) },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			current, _ := tokens.Parse(raw)
			change(current)
			invalid, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, current).SignedString([]byte(cfg.Secret))
			if _, err := tokens.Parse(invalid); err == nil {
				t.Fatal("invalid claims accepted")
			}
		})
	}
	invalid, _ := jwt.NewWithClaims(jwt.SigningMethodHS384, claims).SignedString([]byte(cfg.Secret))
	if _, err := tokens.Parse(invalid); err == nil {
		t.Fatal("wrong algorithm accepted")
	}
	invalid, _ = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(strings.Repeat("x", 48)))
	if _, err := tokens.Parse(invalid); err == nil {
		t.Fatal("wrong key accepted")
	}
	for _, value := range []string{"", strings.Repeat("a", 4097), "invalid"} {
		if _, err := tokens.Parse(value); err == nil {
			t.Fatal("malformed token accepted")
		}
	}
}

func TestMemberAndAdminKindsAreIsolated(t *testing.T) {
	// Even accidental reuse of a signing key/issuer/audience cannot bridge identity kinds.
	c := Config{Secret: strings.Repeat("k", 48), Issuer: "shared-test", Audience: "shared-test", AccessTTL: 15 * time.Minute, Kind: "admin"}
	admin, err := New(c)
	if err != nil {
		t.Fatal(err)
	}
	c.Kind = "member"
	member, err := New(c)
	if err != nil {
		t.Fatal(err)
	}
	sid, _ := Random()
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
	if _, err = New(c); err == nil {
		t.Fatal("unknown kind accepted")
	}
}
