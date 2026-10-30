package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/hibiken/asynq"
	"github.com/kar1hsu/go-seed/services/system/internal/task"
	"testing"
	"time"
)

func TestCSVFormulaNeutralization(t *testing.T) {
	for _, value := range []string{"=1+1", "+SUM(A1)", "-1+2", "@SUM(A1)", "  =cmd", "\tvalue", "\rvalue"} {
		if got := CSVCell(value); got != "'"+value {
			t.Errorf("unsafe cell %q", got)
		}
	}
	if CSVCell("正常文本") != "正常文本" {
		t.Fatal("ordinary text changed")
	}
}

func TestFutureScheduleIsRetryable(t *testing.T) {
	payload, _ := json.Marshal(task.Schedule{Slot: time.Now().UTC().Truncate(time.Hour).Add(time.Hour).Unix()})
	service := &Service{}
	if err := service.Schedule(context.Background(), payload); err == nil || errors.Is(err, asynq.SkipRetry) {
		t.Fatal("clock skew must remain retryable", err)
	}
	if err := service.Schedule(context.Background(), []byte("invalid")); !errors.Is(err, asynq.SkipRetry) {
		t.Fatal("malformed payload should not retry", err)
	}
}
