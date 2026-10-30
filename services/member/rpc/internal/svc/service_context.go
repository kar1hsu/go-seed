package svc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/rpcx"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/member/internal/application/account"
	"github.com/kar1hsu/go-seed/services/member/internal/application/management"
	"github.com/kar1hsu/go-seed/services/member/internal/repository"
	"github.com/kar1hsu/go-seed/services/member/rpc/internal/config"
	systempb "github.com/kar1hsu/go-seed/services/system/rpc/pb"
	systemclient "github.com/kar1hsu/go-seed/services/system/rpc/system"
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type ServiceContext struct {
	Config     config.Config
	DB         *gorm.DB
	Cache      *redis.Redis
	Account    *account.Service
	Management *management.Service
	System     systemclient.System
	client     zrpc.Client
}
type accessChecker struct{ system systemclient.System }

func (a accessChecker) Check(ctx context.Context, permission string) (uint64, error) {
	reply, err := a.system.CheckAccess(ctx, &systempb.CheckAccessRequest{Permission: permission})
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseUint(reply.AdminId, 10, 64)
	if err != nil || id == 0 {
		return 0, xerr.New(codes.Internal)
	}
	return id, nil
}
func NewServiceContext(c config.Config) (*ServiceContext, error) {
	if c.Tokens.Kind != "member" {
		return nil, errors.New("member token kind required")
	}
	tokens, err := identity.New(c.Tokens)
	if err != nil {
		return nil, err
	}
	interceptor, err := rpcx.Client(c.SystemCredential)
	if err != nil {
		return nil, err
	}
	c.SystemRPC.Middlewares.Duration = false
	client, err := zrpc.NewClient(c.SystemRPC, zrpc.WithUnaryClientInterceptor(interceptor))
	if err != nil {
		return nil, err
	}
	db, err := dbx.OpenMySQL(c.MySQL)
	if err != nil {
		client.Conn().Close()
		return nil, err
	}
	pool, err := db.DB()
	if err != nil {
		client.Conn().Close()
		return nil, err
	}
	cleanup := func() { pool.Close(); client.Conn().Close() }
	cache, err := redis.NewRedis(c.Cache)
	if err != nil {
		cleanup()
		return nil, err
	}
	prefix := "goseed:" + c.Mode + ":member:auth:"
	registerLimit := limit.NewPeriodLimit(60, c.RegisterPerMinute, cache, prefix+"register:")
	loginLimit := limit.NewPeriodLimit(60, c.LoginPerMinute, cache, prefix+"login-global:")
	accountLimit := limit.NewPeriodLimit(60, 10, cache, prefix+"login-account:")
	take := func(ctx context.Context, l *limit.PeriodLimit, key string) error {
		result, err := l.TakeCtx(ctx, key)
		if err != nil {
			return xerr.New(codes.Unavailable)
		}
		if result == limit.OverQuota {
			return xerr.New(codes.ResourceExhausted)
		}
		return nil
	}
	auth, err := account.New(repository.NewAuth(db), tokens, c.RefreshTTL, func(ctx context.Context, key string) error {
		if strings.HasPrefix(key, "register:") {
			return take(ctx, registerLimit, "all")
		}
		if err := take(ctx, loginLimit, "all"); err != nil {
			return err
		}
		digest := sha256.Sum256([]byte(key))
		return take(ctx, accountLimit, hex.EncodeToString(digest[:]))
	})
	if err != nil {
		cleanup()
		return nil, err
	}
	system := systemclient.NewSystem(client)
	return &ServiceContext{Config: c, DB: db, Cache: cache, Account: auth, Management: management.New(repository.NewAuth(db), accessChecker{system}), System: system, client: client}, nil
}
func (s *ServiceContext) Ready(ctx context.Context) error {
	pool, err := s.DB.DB()
	if err != nil {
		return err
	}
	if err = pool.PingContext(ctx); err != nil {
		return err
	}
	if !s.Cache.PingCtx(ctx) {
		return errors.New("member redis unavailable")
	}
	_, err = s.System.CheckReady(ctx, &systempb.CheckReadyRequest{})
	return err
}
func (s *ServiceContext) Close() error {
	pool, err := s.DB.DB()
	if err != nil {
		return err
	}
	return errors.Join(pool.Close(), s.client.Conn().Close())
}
