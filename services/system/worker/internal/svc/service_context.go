package svc

import (
	"context"
	"github.com/hibiken/asynq"
	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/pkg/storage"
	"github.com/kar1hsu/go-seed/pkg/taskx"
	"github.com/kar1hsu/go-seed/services/system/internal/application/jobs"
	"github.com/kar1hsu/go-seed/services/system/internal/authorization"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/worker/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"time"
)

type ServiceContext struct {
	DB      *gorm.DB
	Objects storage.Store
	Jobs    *jobs.Service
	Client  *asynq.Client
	Server  *asynq.Server
	Redis   *redis.Client
}

func New(c config.Config) (*ServiceContext, error) {
	if err := c.Queue.Validate(); err != nil {
		return nil, err
	}
	objects, err := storage.New(c.Storage)
	if err != nil {
		return nil, err
	}
	db, err := dbx.OpenMySQL(c.MySQL)
	if err != nil {
		return nil, err
	}
	r := redis.NewClient(&redis.Options{Addr: c.Queue.Addr, Password: c.Queue.Password, DB: c.Queue.DB})
	jobsSvc := jobs.New(repository.NewFilesTasks(db), repository.NewRBAC(db), authorization.NewDeferred(), objects)
	jobsSvc.SetAuditRetention(c.AuditRetentionDays)
	return &ServiceContext{DB: db, Objects: objects, Jobs: jobsSvc, Client: asynq.NewClient(c.Queue.Redis()), Server: asynq.NewServer(c.Queue.Redis(), asynq.Config{ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, t *asynq.Task, err error) {
		id, _ := asynq.GetTaskID(ctx)
		logx.WithContext(ctx).Errorw("queue execution failed", logx.Field("taskId", id), logx.Field("type", t.Type()))
	}), Concurrency: c.Concurrency, Queues: map[string]int{c.Queue.Name: 1}, ShutdownTimeout: taskx.WorkerShutdownTimeout, RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
		return time.Duration(min(1<<min(n, 6), 60)) * time.Second
	}}), Redis: r}, nil
}
func (s *ServiceContext) Ready(ctx context.Context) error {
	db, err := s.DB.DB()
	if err != nil {
		return err
	}
	if err = db.PingContext(ctx); err != nil {
		return err
	}
	if err = s.Redis.Ping(ctx).Err(); err != nil {
		return err
	}
	return s.Objects.Check(ctx)
}
func (s *ServiceContext) Close() {
	s.Client.Close()
	s.Redis.Close()
	if db, err := s.DB.DB(); err == nil {
		db.Close()
	}
}
