// Package management implements administrator use cases over member-owned data.
package management

import (
	"context"
	"errors"
	"fmt"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/member/internal/application/account"
	"github.com/kar1hsu/go-seed/services/member/internal/model"
	"github.com/kar1hsu/go-seed/services/member/internal/repository"
	"github.com/kar1hsu/go-seed/services/member/resources"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

// AccessChecker is implemented by the RPC adapter; use cases never import another domain's internals.
type AccessChecker interface {
	Check(context.Context, string) (uint64, error)
}
type Service struct {
	store  *repository.Auth
	access AccessChecker
}
type Query struct {
	Page, PageSize   int64
	Keyword, Enabled string
}

func New(store *repository.Auth, access AccessChecker) *Service { return &Service{store, access} }
func ID(raw string) (uint64, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 || strconv.FormatUint(id, 10) != raw {
		return 0, xerr.New(codes.InvalidArgument)
	}
	return id, nil
}
func (s *Service) List(ctx context.Context, q Query) ([]model.User, int64, error) {
	if q.Page < 1 || q.Page > 100000 || q.PageSize < 1 || q.PageSize > 100 || len([]rune(q.Keyword)) > 64 || (q.Enabled != "all" && q.Enabled != "enabled" && q.Enabled != "disabled") {
		return nil, 0, xerr.New(codes.InvalidArgument)
	}
	if _, err := s.access.Check(ctx, resources.UserRead); err != nil {
		return nil, 0, err
	}
	users, total, err := s.store.List(ctx, q.Page, q.PageSize, strings.TrimSpace(q.Keyword), q.Enabled)
	return users, total, publicError(ctx, err)
}
func (s *Service) Get(ctx context.Context, id uint64) (*model.User, []model.AuditEvent, error) {
	if id == 0 {
		return nil, nil, xerr.New(codes.InvalidArgument)
	}
	if _, err := s.access.Check(ctx, resources.UserRead); err != nil {
		return nil, nil, err
	}
	u, events, err := s.store.Detail(ctx, id)
	return u, events, publicError(ctx, err)
}
func (s *Service) SetStatus(ctx context.Context, id uint64, enabled bool, revision string) error {
	if id == 0 || revision == "" {
		return xerr.New(codes.InvalidArgument)
	}
	// Authorization completes before starting this domain's transaction. In-flight authorized operations may complete.
	actor, err := s.access.Check(ctx, resources.UserUpdateStatus)
	if err != nil {
		return err
	}
	return publicError(ctx, s.store.Transaction(ctx, func(tx *repository.Auth) error {
		user, err := tx.LockUser(ctx, id)
		if err != nil {
			return err
		}
		if account.Revision(user.UpdatedAt) != revision {
			return xerr.New(codes.Aborted)
		}
		if user.Enabled == enabled {
			return nil
		}
		if err = tx.Status(ctx, id, enabled, account.NextTime(user.UpdatedAt)); err != nil {
			return err
		}
		action := "user.disable"
		if enabled {
			action = "user.enable"
		}
		return tx.Event(ctx, id, action, "admin", actor)
	}))
}
func publicError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return xerr.New(codes.NotFound)
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	logx.WithContext(ctx).Error(fmt.Sprintf("member storage failure (%T)", err))
	return xerr.New(codes.Unavailable)
}
