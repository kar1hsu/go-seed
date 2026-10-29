package health

import (
	"context"
	"fmt"
	"time"
)

type PingFunc func(context.Context) error

type Service struct{ mysql, redis PingFunc }

func New(mysql, redis PingFunc) *Service { return &Service{mysql: mysql, redis: redis} }

func (s *Service) Ready(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()
	if err := s.mysql(ctx); err != nil {
		return fmt.Errorf("mysql readiness: %w", err)
	}
	if err := s.redis(ctx); err != nil {
		return fmt.Errorf("redis readiness: %w", err)
	}
	return nil
}
