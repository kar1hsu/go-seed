package main

import (
	"context"
	"flag"
	"github.com/kar1hsu/go-seed/pkg/taskx"
	"github.com/kar1hsu/go-seed/pkg/timezone"
	"github.com/kar1hsu/go-seed/services/system/scheduler/internal/config"
	"github.com/kar1hsu/go-seed/services/system/scheduler/internal/schedule"
	"github.com/kar1hsu/go-seed/services/system/scheduler/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := timezone.Configure(); err != nil {
		log.Fatal(err)
	}
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	path := flag.String("f", "services/system/scheduler/etc/system-scheduler.yaml", "configuration")
	flag.Parse()
	var c config.Config
	conf.MustLoad(*path, &c, conf.UseEnv())
	if err := c.SetUp(); err != nil {
		return err
	}
	s, err := svc.New(c)
	if err != nil {
		return err
	}
	defer s.Close()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	health := taskx.HealthServer(c.ListenOn, s.Ready)
	var listenerError error
	failed := make(chan error, 1)
	go func() { failed <- health.ListenAndServe() }()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		call, stop := context.WithTimeout(ctx, 3*time.Second)
		err = schedule.Submit(call, s.Client, c.Queue.Name, time.Now())
		stop()
		if err != nil && ctx.Err() == nil {
			logx.Errorw("scheduler delivery unavailable; retrying current hourly slot")
		}
		select {
		case <-ctx.Done():
		case err = <-failed:
			if err != http.ErrServerClosed {
				listenerError = err
			}
			cancel()
		case <-ticker.C:
			continue
		}
		break
	}
	stopCtx, stop := context.WithTimeout(context.Background(), 3*time.Second)
	defer stop()
	_ = health.Shutdown(stopCtx)
	return listenerError
}
