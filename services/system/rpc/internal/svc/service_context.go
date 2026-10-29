package svc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/internal/application/auth"
	"github.com/kar1hsu/go-seed/services/system/internal/application/health"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"
	"github.com/kar1hsu/go-seed/services/system/internal/authorization"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/config"
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config     config.Config
	DB         *gorm.DB
	Cache      *redis.Redis
	Health     *health.Service
	Auth       *auth.Service
	Management *management.Service
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	tokens, err := identity.New(c.Tokens)
	if err != nil {
		return nil, err
	}
	db, err := dbx.OpenMySQL(c.MySQL)
	if err != nil {
		return nil, err
	}
	pool, err := db.DB()
	if err != nil {
		return nil, err
	}
	cache, err := redis.NewRedis(c.Cache)
	if err != nil {
		_ = pool.Close()
		return nil, err
	}
	healthSvc := health.New(pool.PingContext, func(ctx context.Context) error {
		if !cache.PingCtx(ctx) {
			return errors.New("redis ping failed")
		}
		return nil
	})
	loginLimit := limit.NewPeriodLimit(60, 10, cache, "goseed:"+c.Mode+":system:auth:login:")
	authSvc, err := auth.New(repository.NewAuth(db), tokens, c.RefreshTTL, func(ctx context.Context, username string) error {
		key := sha256.Sum256([]byte(username))
		result, err := loginLimit.TakeCtx(ctx, hex.EncodeToString(key[:]))
		if err != nil {
			return xerr.New(codes.Unavailable)
		}
		if result == limit.OverQuota {
			return xerr.New(codes.ResourceExhausted)
		}
		return nil
	})
	if err != nil {
		_ = pool.Close()
		return nil, err
	}
	managementSvc, err := management.New(repository.NewRBAC(db), authorization.New(tokens))
	if err != nil {
		_ = pool.Close()
		return nil, err
	}
	return &ServiceContext{Config: c, DB: db, Cache: cache, Health: healthSvc, Auth: authSvc, Management: managementSvc}, nil
}

func (s *ServiceContext) Close() error {
	pool, err := s.DB.DB()
	if err != nil {
		return err
	}
	return pool.Close()
}
