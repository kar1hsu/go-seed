// Package authlimit separates anonymous source quotas from authenticated password attempts.
package authlimit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/kar1hsu/go-seed/pkg/clientip"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/grpc/codes"
	"strings"
)

type Take func(context.Context, string) error
type Limiter struct{ Login, Account, Register, Password Take }

func New(cache *redis.Redis, prefix string, login, register int) *Limiter {
	makeLimit := func(name string, n int) Take {
		l := limit.NewPeriodLimit(60, n, cache, prefix+name+":")
		return func(ctx context.Context, key string) error {
			digest := sha256.Sum256([]byte(key))
			result, err := l.TakeCtx(ctx, hex.EncodeToString(digest[:]))
			if err != nil {
				return xerr.New(codes.Unavailable)
			}
			if result == limit.OverQuota {
				return xerr.New(codes.ResourceExhausted)
			}
			return nil
		}
	}
	return &Limiter{makeLimit("login-ip", login), makeLimit("login-ip-account", 10), makeLimit("register-ip", register), makeLimit("password-account", 5)}
}
func (l *Limiter) Allow(ctx context.Context, key string) error {
	// This key is produced only after validating the authenticated account, never from request body.
	if strings.HasPrefix(key, "password:") {
		return l.Password(ctx, key)
	}
	ip := clientip.From(ctx)
	if !clientip.Valid(ip) {
		return xerr.New(codes.Unauthenticated)
	}
	if strings.HasPrefix(key, "register:") {
		return l.Register(ctx, ip)
	}
	if err := l.Login(ctx, ip); err != nil {
		return err
	}
	return l.Account(ctx, ip+"/"+key)
}

// Gate bounds expensive password work without an unbounded waiting queue.
type Gate chan struct{}

func NewGate(n int) Gate { return make(Gate, n) }
func (g Gate) Acquire(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case g <- struct{}{}:
		return func() { <-g }, nil
	default:
		return nil, xerr.New(codes.ResourceExhausted)
	}
}
