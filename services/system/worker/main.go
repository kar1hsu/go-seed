package main

import (
	"context"
	"flag"
	"github.com/kar1hsu/go-seed/pkg/taskx"
	"github.com/kar1hsu/go-seed/services/system/worker/internal/config"
	"github.com/kar1hsu/go-seed/services/system/worker/internal/handler"
	"github.com/kar1hsu/go-seed/services/system/worker/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	path := flag.String("f", "services/system/worker/etc/system-worker.yaml", "configuration")
	flag.Parse()
	var c config.Config
	conf.MustLoad(*path, &c, conf.UseEnv())
	if err := c.SetUp(); err != nil {
		log.Fatal(err)
	}
	s, err := svc.New(c)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	if err = s.Server.Start(handler.New(s.Jobs)); err != nil {
		log.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			call, stop := context.WithTimeout(ctx, 5*time.Second)
			err := s.Jobs.Dispatch(call, s.Client.EnqueueContext, c.Queue.Name)
			stop()
			if err != nil && ctx.Err() == nil {
				logx.Errorw("outbox delivery unavailable; records remain pending")
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	health := taskx.HealthServer(c.ListenOn, s.Ready)
	failed := make(chan error, 1)
	go func() { failed <- health.ListenAndServe() }()
	select {
	case <-ctx.Done():
	case err = <-failed:
		if err != http.ErrServerClosed {
			logx.Errorf("worker health listener failed: %v", err)
		}
	}
	cancel()
	s.Server.Shutdown()
	<-done
	stopCtx, stop := context.WithTimeout(context.Background(), 3*time.Second)
	defer stop()
	_ = health.Shutdown(stopCtx)
}
