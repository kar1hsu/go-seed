// Package identity contains transport-neutral token validation, not user persistence.
package identity

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Config contains only public verification material. API processes use this type.
type Config struct {
	Kind      string `json:",default=admin"`
	PublicKey string
	Issuer    string
	Audience  string
}

// SignerConfig belongs only to the issuing domain. RefreshKey is an independent
// HMAC key; rotating a JWT key must not change refresh recovery derivation.
type SignerConfig struct {
	Config
	PrivateKey string
	RefreshKey string
	AccessTTL  time.Duration `json:",default=15m"`
}

type Claims struct {
	jwt.RegisteredClaims
	Kind      string `json:"kind"`
	SessionID string `json:"sid"`
	Version   uint64 `json:"ver"`
}

type Verifier struct {
	config    Config
	publicKey ed25519.PublicKey
}

type Signer struct {
	*Verifier
	privateKey ed25519.PrivateKey
	refreshKey []byte
	accessTTL  time.Duration
}

func decodeKey(value string, size int) ([]byte, error) {
	key, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || len(key) != size || base64.RawURLEncoding.EncodeToString(key) != value {
		return nil, errors.New("invalid key encoding or length")
	}
	return key, nil
}

func NewVerifier(c Config) (*Verifier, error) {
	if c.Kind == "" {
		c.Kind = "admin"
	}
	if c.Kind != "admin" && c.Kind != "member" || c.Issuer == "" || c.Audience == "" {
		return nil, errors.New("invalid token verification configuration")
	}
	public, err := decodeKey(c.PublicKey, ed25519.PublicKeySize)
	if err != nil {
		return nil, errors.New("invalid Ed25519 public key")
	}
	return &Verifier{config: c, publicKey: ed25519.PublicKey(public)}, nil
}

func NewSigner(c SignerConfig) (*Signer, error) {
	verifier, err := NewVerifier(c.Config)
	if err != nil {
		return nil, err
	}
	private, err := decodeKey(c.PrivateKey, ed25519.PrivateKeySize)
	if err != nil {
		return nil, errors.New("invalid Ed25519 private key")
	}
	// Check both the seed and the cached public half, rather than trusting a
	// malformed 64-byte key whose public half was supplied independently.
	derived := ed25519.NewKeyFromSeed(private[:ed25519.SeedSize])
	if subtle.ConstantTimeCompare(private, derived) != 1 || subtle.ConstantTimeCompare(derived[ed25519.SeedSize:], verifier.publicKey) != 1 {
		return nil, errors.New("Ed25519 key pair does not match")
	}
	refresh, err := decodeKey(c.RefreshKey, 32)
	if err != nil || subtle.ConstantTimeCompare(refresh, private[:ed25519.SeedSize]) == 1 || subtle.ConstantTimeCompare(refresh, verifier.publicKey) == 1 {
		return nil, errors.New("refresh recovery requires an independent 32-byte key")
	}
	if c.AccessTTL < time.Minute || c.AccessTTL > time.Hour {
		return nil, errors.New("invalid access token lifetime")
	}
	return &Signer{Verifier: verifier, privateKey: ed25519.PrivateKey(private), refreshKey: refresh, accessTTL: c.AccessTTL}, nil
}

func Random() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (t *Signer) Issue(id uint64, sid string, version uint64) (string, time.Time, error) {
	now := time.Now().UTC()
	expires := now.Add(t.accessTTL).Truncate(time.Second)
	jti, err := Random()
	if err != nil {
		return "", time.Time{}, err
	}
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    t.config.Issuer,
			Subject:   strconv.FormatUint(id, 10),
			Audience:  jwt.ClaimStrings{t.config.Audience},
			ExpiresAt: jwt.NewNumericDate(expires),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        jti,
		},
		Kind:      t.config.Kind,
		SessionID: sid,
		Version:   version,
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(t.privateKey)
	return raw, expires, err
}

func (t *Verifier) Parse(raw string) (*Claims, error) {
	if len(raw) > 4096 || raw == "" {
		return nil, errors.New("invalid access token")
	}
	claims := new(Claims)
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"EdDSA"}))
	parsed, err := parser.ParseWithClaims(raw, claims, func(_ *jwt.Token) (any, error) {
		return t.publicKey, nil
	})
	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid access token")
	}
	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil || id == 0 || claims.Kind != t.config.Kind || len(claims.SessionID) != 43 || claims.Version == 0 || claims.ID == "" {
		return nil, errors.New("invalid access claims")
	}
	if claims.ExpiresAt == nil || claims.IssuedAt == nil || claims.NotBefore == nil {
		return nil, errors.New("missing token timestamps")
	}
	if !claims.VerifyIssuer(t.config.Issuer, true) || !claims.VerifyAudience(t.config.Audience, true) {
		return nil, errors.New("invalid access claims")
	}
	return claims, nil
}

func Bearer(value string) string {
	fields := strings.Fields(value)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return ""
	}
	return fields[1]
}

type tokenKey struct{}

func WithToken(ctx context.Context, raw string) context.Context {
	return context.WithValue(ctx, tokenKey{}, raw)
}

func RawToken(ctx context.Context) string {
	value, _ := ctx.Value(tokenKey{}).(string)
	return value
}
