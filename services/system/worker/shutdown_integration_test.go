//go:build integration

package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/taskx"
	"github.com/kar1hsu/go-seed/services/system/worker/internal/config"
	"github.com/zeromicro/go-zero/core/conf"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"
)

// A real subprocess installs go-zero's signal hooks using the worker YAML, then
// receives SIGTERM during an Asynq job longer than the old 5.5s force-quit window.
func TestWorkerGracefulStop(t *testing.T) {
	if os.Getenv("GOSEED_SHUTDOWN_HELPER") == "1" {
		var c config.Config
		if err := conf.Load("etc/system-worker.yaml", &c, conf.UseEnv()); err != nil {
			t.Fatal(err)
		}
		if err := c.SetUp(); err != nil {
			t.Fatal(err)
		}
		server := asynq.NewServer(c.Queue.Redis(), asynq.Config{Concurrency: 1, Queues: map[string]int{c.Queue.Name: 1}, ShutdownTimeout: taskx.WorkerShutdownTimeout})
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
		defer cancel()
		if err := server.Start(asynq.HandlerFunc(func(context.Context, *asynq.Task) error {
			fmt.Println("GOSEED_JOB_STARTED")
			time.Sleep(6 * time.Second)
			fmt.Println("GOSEED_JOB_DONE")
			return nil
		})); err != nil {
			t.Fatal(err)
		}
		<-ctx.Done()
		server.Shutdown()
		return
	}
	if os.Getenv("QUEUE_PORT") == "" {
		t.Skip("run make test-rbac with project Redis")
	}
	var c config.Config
	if err := conf.Load("etc/system-worker.yaml", &c, conf.UseEnv()); err != nil {
		t.Fatal(err)
	}
	key, err := identity.Random()
	if err != nil {
		t.Fatal(err)
	}
	queue := "goseed_shutdown_test_" + key
	inspector := asynq.NewInspector(c.Queue.Redis())
	defer func() { inspector.DeleteQueue(queue, true); inspector.Close() }()
	client := asynq.NewClient(c.Queue.Redis())
	defer client.Close()
	if _, err := client.Enqueue(asynq.NewTask("test:graceful-stop", nil), asynq.Queue(queue), asynq.MaxRetry(0)); err != nil {
		t.Fatal(err)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "-test.run=^TestWorkerGracefulStop$", "-test.timeout=20s")
	cmd.Env = append(os.Environ(), "GOSEED_SHUTDOWN_HELPER=1", "QUEUE_NAME="+queue)
	output, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = cmd.Stdout
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	started, finished := make(chan struct{}), make(chan struct{})
	go func() {
		scanner := bufio.NewScanner(output)
		for scanner.Scan() {
			switch strings.TrimSpace(scanner.Text()) {
			case "GOSEED_JOB_STARTED":
				close(started)
			case "GOSEED_JOB_DONE":
				close(finished)
			}
		}
	}()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("worker did not start task")
	}
	if err = cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if err = cmd.Wait(); err != nil {
		t.Fatal("worker forced to exit before completing in-flight task", err)
	}
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("in-flight task interrupted")
	}
}
