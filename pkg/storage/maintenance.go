package storage

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
)

// Maintain cleans the producer's own spool, including files left by a previous crash.
// It never touches recent files: all upload handlers have deadlines well below 24h.
func Maintain(store Store) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		timer := time.NewTicker(time.Hour)
		defer timer.Stop()
		for {
			call, stop := context.WithTimeout(ctx, 10*time.Second)
			err := store.CleanupTemporary(call, time.Now().UTC().Add(-24*time.Hour))
			stop()
			if err != nil && ctx.Err() == nil {
				logx.Errorw("upload spool cleanup failed")
			}
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
			}
		}
	}()
	return func() { cancel(); <-done }
}
