package handler

import (
	"context"
	"github.com/hibiken/asynq"
	"github.com/kar1hsu/go-seed/services/system/internal/application/jobs"
	"github.com/kar1hsu/go-seed/services/system/internal/task"
)

func New(jobs *jobs.Service) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(task.Execute, func(ctx context.Context, t *asynq.Task) error {
		p, err := task.Decode(t.Payload())
		if err != nil {
			return asynq.SkipRetry
		}
		return jobs.Process(ctx, p)
	})
	mux.HandleFunc(task.Tick, func(ctx context.Context, t *asynq.Task) error { return jobs.Schedule(ctx, t.Payload()) })
	return mux
}
