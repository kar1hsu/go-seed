package svc

import (
	"context"
	"errors"
	"github.com/kar1hsu/go-seed/pkg/authlimit"
	"github.com/kar1hsu/go-seed/pkg/storage"
	"regexp"
	"time"

	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/auth"
	"github.com/kar1hsu/go-seed/services/system/internal/application/health"
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"
	"github.com/kar1hsu/go-seed/services/system/internal/authorization"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
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
	for _, caller := range c.DomainCallers {
		if !regexp.MustCompile(`^[a-z][a-z0-9_]*\.$`).MatchString(caller.PermissionPrefix) || caller.PermissionPrefix == "system." {
			return nil, errors.New("invalid domain permission scope")
		}
	}
	objects, err := storage.New(c.Storage)
	if err != nil {
		return nil, err
	}
	tokens, err := identity.NewSigner(c.Tokens)
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
		if err := objects.Check(ctx); err != nil {
			return err
		}
		if !cache.PingCtx(ctx) {
			return errors.New("redis ping failed")
		}
		return nil
	})
	limiter := authlimit.New(cache, "goseed:"+c.Mode+":system:auth:", 60, 10)
	authSvc, err := auth.New(repository.NewAuth(db), tokens, c.RefreshTTL, limiter.Allow)
	if err != nil {
		_ = pool.Close()
		return nil, err
	}
	managementSvc, err := management.New(repository.NewRBAC(db), authorization.New(tokens.Verifier), objects, time.Local)
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
