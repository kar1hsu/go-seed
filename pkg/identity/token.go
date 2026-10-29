// Package identity contains transport-neutral token validation, not user persistence.
package identity

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Config struct {
	Secret    string
	Issuer    string
	Audience  string
	AccessTTL time.Duration `json:",default=15m"`
}

type Claims struct {
	jwt.RegisteredClaims
	Kind      string `json:"kind"`
	SessionID string `json:"sid"`
	Version   uint64 `json:"ver"`
}

type Tokens struct {
	config Config
}

func New(c Config) (*Tokens, error) {
	if len(c.Secret) < 32 || c.Issuer == "" || c.Audience == "" || c.AccessTTL < time.Minute || c.AccessTTL > time.Hour {
		return nil, errors.New("invalid token configuration")
	}
	return &Tokens{config: c}, nil
}

func Random() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (t *Tokens) Issue(id uint64, sid string, version uint64) (string, time.Time, error) {
	now := time.Now().UTC()
	expires := now.Add(t.config.AccessTTL).Truncate(time.Second)
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
		Kind:      "admin",
		SessionID: sid,
		Version:   version,
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(t.config.Secret))
	return raw, expires, err
}

func (t *Tokens) Parse(raw string) (*Claims, error) {
	if len(raw) > 4096 || raw == "" {
		return nil, errors.New("invalid access token")
	}
	claims := new(Claims)
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"HS256"}))
	parsed, err := parser.ParseWithClaims(raw, claims, func(_ *jwt.Token) (any, error) {
		return []byte(t.config.Secret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid access token")
	}
	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil || id == 0 || claims.Kind != "admin" || len(claims.SessionID) != 43 || claims.Version == 0 || claims.ID == "" {
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
