// Package schedule submits the current UTC hourly slot. Redis outages are retried;
// completed slots are deduplicated again by the worker's database unique key.
package schedule

import (
	"context"
	"errors"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/kar1hsu/go-seed/services/system/internal/task"
	"time"
)

func Submit(ctx context.Context, client *asynq.Client, queue string, now time.Time) error {
	slot := now.UTC().Truncate(time.Hour)
	_, err := client.EnqueueContext(ctx, task.TickMessage(slot), asynq.Queue(queue), asynq.TaskID(fmt.Sprintf("cleanup-tick-%d", slot.Unix())), asynq.MaxRetry(100), asynq.Timeout(10*time.Second), asynq.Retention(48*time.Hour))
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		return nil
	}
	return err
}
