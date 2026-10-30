package svc

import (
	"context"
	"github.com/hibiken/asynq"
	"github.com/kar1hsu/go-seed/services/system/scheduler/internal/config"
	"github.com/redis/go-redis/v9"
)

type ServiceContext struct {
	Client *asynq.Client
	Redis  *redis.Client
}

func New(c config.Config) (*ServiceContext, error) {
	if err := c.Queue.Validate(); err != nil {
		return nil, err
	}
	return &ServiceContext{Client: asynq.NewClient(c.Queue.Redis()), Redis: redis.NewClient(&redis.Options{Addr: c.Queue.Addr, Password: c.Queue.Password, DB: c.Queue.DB})}, nil
}
func (s *ServiceContext) Ready(ctx context.Context) error { return s.Redis.Ping(ctx).Err() }
func (s *ServiceContext) Close()                          { s.Client.Close(); s.Redis.Close() }
