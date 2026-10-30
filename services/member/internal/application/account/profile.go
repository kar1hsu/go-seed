package account

import (
	"context"
	"errors"
	mysql "github.com/go-sql-driver/mysql"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/member/internal/model"
	"github.com/kar1hsu/go-seed/services/member/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func ValidNickname(v string) bool {
	return utf8.ValidString(v) && strings.TrimSpace(v) != "" && len([]rune(v)) <= 64 && !strings.ContainsFunc(v, unicode.IsControl)
}
func Revision(t time.Time) string { return t.UTC().Format(time.RFC3339Nano) }
func NextTime(t time.Time) time.Time {
	now := time.Now().UTC().Truncate(time.Microsecond)
	if !now.After(t) {
		return t.Add(time.Microsecond)
	}
	return now
}
func (s *Service) Register(ctx context.Context, username, password, nickname string) (*model.User, error) {
	if !ValidUsername(username) || !ValidPassword(password) || !ValidNickname(nickname) {
		return nil, xerr.New(codes.InvalidArgument)
	}
	if err := s.allowLogin(ctx, "register:"+username); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), passwordCost)
	if err != nil {
		return nil, storageError(ctx, err)
	}
	now := NextTime(time.Time{})
	u := model.User{Username: username, Nickname: strings.TrimSpace(nickname), PasswordHash: string(hash), Enabled: true, TokenVersion: 1, CreatedAt: now, UpdatedAt: now}
	err = s.store.Transaction(ctx, func(tx *repository.Auth) error {
		if err := tx.CreateUser(ctx, &u); err != nil {
			return err
		}
		return tx.Event(ctx, u.ID, "register", "member", u.ID)
	})
	var duplicate *mysql.MySQLError
	if errors.As(err, &duplicate) && duplicate.Number == 1062 {
		return nil, xerr.New(codes.AlreadyExists)
	}
	return &u, storageError(ctx, err)
}
func (s *Service) UpdateProfile(ctx context.Context, raw, nickname, revision string) (*model.User, error) {
	if !ValidNickname(nickname) || revision == "" {
		return nil, xerr.New(codes.InvalidArgument)
	}
	a, err := s.Current(ctx, raw)
	if err != nil {
		return nil, err
	}
	claims, err := s.tokens.Parse(raw)
	if err != nil {
		return nil, xerr.New(codes.Unauthenticated)
	}
	var result *model.User
	err = s.store.Transaction(ctx, func(tx *repository.Auth) error {
		current, err := tx.LockUser(ctx, a.ID)
		if err != nil {
			return err
		}
		if _, err = tx.Current(ctx, a.ID, claims.SessionID, claims.Version, time.Now().UTC()); err != nil {
			return err
		}
		if revision != Revision(current.UpdatedAt) {
			return xerr.New(codes.Aborted)
		}
		current.Nickname = strings.TrimSpace(nickname)
		current.UpdatedAt = NextTime(current.UpdatedAt)
		if err = tx.Profile(ctx, a.ID, current.Nickname, current.UpdatedAt); err != nil {
			return err
		}
		if err = tx.Event(ctx, a.ID, "profile.update", "member", a.ID); err != nil {
			return err
		}
		result = current
		return nil
	})
	return result, storageError(ctx, err)
}
