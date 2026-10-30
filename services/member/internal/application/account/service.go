package account

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/member/internal/model"
	"github.com/kar1hsu/go-seed/services/member/internal/repository"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

const passwordCost = 12

// usernamePattern is immutable; account identifiers are canonical lowercase ASCII.
var usernamePattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,31}$`)

type Service struct {
	store      *repository.Auth
	tokens     *identity.Tokens
	refreshTTL time.Duration
	allowLogin func(context.Context, string) error
	dummyHash  []byte
}

type Session struct {
	User             *model.User
	AccessToken      string
	ExpiresAt        time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}

func New(store *repository.Auth, tokens *identity.Tokens, refreshTTL time.Duration, allowLogin func(context.Context, string) error) (*Service, error) {
	if refreshTTL < time.Hour || refreshTTL > 30*24*time.Hour || allowLogin == nil {
		return nil, errors.New("invalid authentication configuration")
	}
	dummy, err := bcrypt.GenerateFromPassword([]byte("invalid-account-placeholder"), passwordCost)
	if err != nil {
		return nil, err
	}
	return &Service{store: store, tokens: tokens, refreshTTL: refreshTTL, allowLogin: allowLogin, dummyHash: dummy}, nil
}

func ValidUsername(value string) bool { return usernamePattern.MatchString(value) }
func ValidPassword(value string) bool {
	return len(value) >= 12 && len(value) <= 72 && !strings.ContainsRune(value, 0)
}

func (s *Service) Login(ctx context.Context, username, password string) (*Session, error) {
	if !ValidUsername(username) || len(password) == 0 || len(password) > 72 {
		return nil, xerr.New(codes.InvalidArgument)
	}
	if err := s.allowLogin(ctx, username); err != nil {
		return nil, err
	}
	a, err := s.store.ByUsername(ctx, username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(password))
		return nil, xerr.New(codes.Unauthenticated)
	}
	if err != nil {
		return nil, storageError(ctx, err)
	}
	if bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password)) != nil || !a.Enabled {
		return nil, xerr.New(codes.Unauthenticated)
	}
	var result *Session
	err = s.store.Transaction(ctx, func(tx *repository.Auth) error {
		current, err := tx.LockUser(ctx, a.ID)
		if err != nil {
			return err
		}
		if !current.Enabled || current.PasswordHash != a.PasswordHash {
			return xerr.New(codes.Unauthenticated)
		}
		sid, err := identity.Random()
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if err := tx.TouchLogin(ctx, current.ID, now); err != nil {
			return err
		}
		session := &model.MemberSession{ID: sid, UserID: current.ID, TokenVersion: current.TokenVersion, ExpiresAt: now.Add(s.refreshTTL), CreatedAt: now}
		if err = tx.CreateSession(ctx, session); err != nil {
			return err
		}
		result, err = s.issue(ctx, tx, current, session)
		return err
	})
	if err != nil {
		return nil, storageError(ctx, err)
	}
	return result, nil
}

func (s *Service) issue(ctx context.Context, tx *repository.Auth, a *model.User, session *model.MemberSession) (*Session, error) {
	secret, err := identity.Random()
	if err != nil {
		return nil, err
	}
	refresh := session.ID + "." + secret
	token := model.MemberRefreshToken{Digest: digest(refresh), SessionID: session.ID, CreatedAt: time.Now().UTC()}
	if err = tx.CreateToken(ctx, &token); err != nil {
		return nil, err
	}
	access, expires, err := s.tokens.Issue(a.ID, session.ID, a.TokenVersion)
	if err != nil {
		return nil, err
	}
	return &Session{User: a, AccessToken: access, ExpiresAt: expires, RefreshToken: refresh, RefreshExpiresAt: session.ExpiresAt}, nil
}

func (s *Service) Refresh(ctx context.Context, raw string) (*Session, error) {
	sid, err := refreshSessionID(raw)
	if err != nil {
		return nil, err
	}
	previous, err := s.store.Session(ctx, sid)
	if err != nil {
		return nil, storageError(ctx, err)
	}
	var result *Session
	var rejected bool
	err = s.store.Transaction(ctx, func(tx *repository.Auth) error {
		// All session mutations lock the member first, including password changes.
		a, err := tx.LockUser(ctx, previous.UserID)
		if err != nil {
			return err
		}
		session, err := tx.Session(ctx, sid)
		if err != nil {
			return err
		}
		if !validSession(a, session, time.Now().UTC()) {
			return xerr.New(codes.Unauthenticated)
		}
		token, err := tx.Token(ctx, digest(raw), sid)
		if err != nil {
			return err
		}
		if token.ConsumedAt != nil {
			// Commit the revocation before returning 401; returning an error inside the transaction would undo it.
			rejected = true
			return tx.Revoke(ctx, sid, time.Now().UTC())
		}
		if err = tx.ConsumeToken(ctx, token.Digest, time.Now().UTC()); err != nil {
			return err
		}
		result, err = s.issue(ctx, tx, a, session)
		return err
	})
	if err != nil {
		return nil, storageError(ctx, err)
	}
	if rejected {
		return nil, xerr.New(codes.Unauthenticated)
	}
	return result, nil
}

func (s *Service) Logout(ctx context.Context, raw string) error {
	if raw == "" {
		return nil
	}
	sid, err := refreshSessionID(raw)
	if err != nil {
		return err
	}
	previous, err := s.store.Session(ctx, sid)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return storageError(ctx, err)
	}
	err = s.store.Transaction(ctx, func(tx *repository.Auth) error {
		if _, err := tx.LockUser(ctx, previous.UserID); err != nil {
			return err
		}
		if _, err := tx.Token(ctx, digest(raw), sid); err != nil {
			return err
		}
		return tx.Revoke(ctx, sid, time.Now().UTC())
	})
	return storageError(ctx, err)
}

func (s *Service) Current(ctx context.Context, raw string) (*model.User, error) {
	claims, err := s.tokens.Parse(raw)
	if err != nil {
		return nil, xerr.New(codes.Unauthenticated)
	}
	id, _ := strconv.ParseUint(claims.Subject, 10, 64)
	a, err := s.store.Current(ctx, id, claims.SessionID, claims.Version, time.Now().UTC())
	return a, storageError(ctx, err)
}

func (s *Service) ChangePassword(ctx context.Context, raw, currentPassword, newPassword string) error {
	if !ValidPassword(newPassword) || len(currentPassword) > 72 || currentPassword == newPassword {
		return xerr.New(codes.InvalidArgument)
	}
	a, err := s.Current(ctx, raw)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(currentPassword)) != nil {
		return xerr.New(codes.InvalidArgument)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), passwordCost)
	if err != nil {
		return err
	}
	claims, err := s.tokens.Parse(raw)
	if err != nil {
		return xerr.New(codes.Unauthenticated)
	}
	err = s.store.Transaction(ctx, func(tx *repository.Auth) error {
		current, err := tx.LockUser(ctx, a.ID)
		if err != nil {
			return err
		}
		if current.PasswordHash != a.PasswordHash {
			return xerr.New(codes.Unauthenticated)
		}
		if _, err = tx.Current(ctx, a.ID, claims.SessionID, claims.Version, time.Now().UTC()); err != nil {
			return err
		}
		if err = tx.Password(ctx, a.ID, string(hash), time.Now().UTC()); err != nil {
			return err
		}
		return tx.Event(ctx, current.ID, "password.change", "member", current.ID)
	})
	return storageError(ctx, err)
}

func validSession(a *model.User, session *model.MemberSession, now time.Time) bool {
	return a.Enabled && a.TokenVersion == session.TokenVersion && session.RevokedAt == nil && session.ExpiresAt.After(now)
}
func refreshSessionID(raw string) (string, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 2 || len(parts[0]) != 43 || len(parts[1]) != 43 {
		return "", xerr.New(codes.Unauthenticated)
	}
	return parts[0], nil
}
func digest(raw string) string {
	value := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(value[:])
}
func storageError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return xerr.New(codes.Unauthenticated)
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	logx.WithContext(ctx).Error(fmt.Sprintf("authentication storage failure (%T)", err))
	return xerr.New(codes.Unavailable)
}
