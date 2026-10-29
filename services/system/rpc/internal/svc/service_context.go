package svc

import (
	"context"
	"errors"

	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/services/system/internal/application/health"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Cache  *redis.Redis
	Health *health.Service
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
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
	return &ServiceContext{Config: c, DB: db, Cache: cache, Health: healthSvc}, nil
}

func (s *ServiceContext) Close() error {
	pool, err := s.DB.DB()
	if err != nil {
		return err
	}
	return pool.Close()
}
