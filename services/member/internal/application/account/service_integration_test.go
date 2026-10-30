//go:build integration

package account_test

import (
	"context"
	"fmt"
	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/member/internal/application/account"
	"github.com/kar1hsu/go-seed/services/member/internal/application/management"
	"github.com/kar1hsu/go-seed/services/member/internal/model"
	"github.com/kar1hsu/go-seed/services/member/internal/repository"
	"github.com/kar1hsu/go-seed/services/member/migrations"
	"github.com/kar1hsu/go-seed/services/member/resources"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const password = "member-test-password-one"

type checker struct{ err error }

func (c checker) Check(_ context.Context, p string) (uint64, error) {
	if p != resources.UserRead && p != resources.UserUpdateStatus {
		return 0, xerr.New(codes.PermissionDenied)
	}
	return 42, c.err
}

type fixture struct {
	db      *gorm.DB
	a, b    *account.Service
	manager *management.Service
}

func setup(t *testing.T) *fixture {
	t.Helper()
	port, err := strconv.Atoi(os.Getenv("MYSQL_PORT"))
	if err != nil {
		t.Fatal("run make test-member")
	}
	cfg := dbx.MySQLConfig{Host: os.Getenv("MYSQL_HOST"), Port: port, Database: "mysql", User: "root", Password: os.Getenv("MYSQL_ROOT_PASSWORD"), MaxOpenConns: 12, MaxIdleConns: 2, ConnMaxLifetime: time.Minute}
	root, err := dbx.OpenMySQL(cfg)
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := root.DB()
	suffix, _ := identity.Random()
	name := "goseed_member_test_" + strings.ToLower(strings.NewReplacer("-", "a", "_", "b").Replace(suffix[:12]))
	if err = root.Exec("CREATE DATABASE `" + name + "` CHARACTER SET utf8mb4").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Exec("DROP DATABASE `" + name + "`"); pool.Close() })
	cfg.Database = name
	db, err := dbx.OpenMySQL(cfg)
	if err != nil {
		t.Fatal(err)
	}
	dbpool, _ := db.DB()
	t.Cleanup(func() { dbpool.Close() })
	for range 2 {
		if err = migrations.Up(context.Background(), dbpool); err != nil {
			t.Fatal(err)
		}
	}
	tokens, err := identity.New(identity.Config{Kind: "member", Secret: strings.Repeat("m", 48), Issuer: "test-member", Audience: "test-app", AccessTTL: 15 * time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	create := func() *account.Service {
		s, err := account.New(repository.NewAuth(db), tokens, 24*time.Hour, func(context.Context, string) error { return nil })
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	return &fixture{db: db, a: create(), b: create(), manager: management.New(repository.NewAuth(db), checker{})}
}
func code(t *testing.T, err error, want codes.Code) {
	t.Helper()
	if status.Code(err) != want {
		t.Fatalf("got %v; want %v", err, want)
	}
}
func register(t *testing.T, f *fixture) *model.User {
	t.Helper()
	u, err := f.a.Register(context.Background(), "sample_member", password, "样板会员")
	if err != nil {
		t.Fatal(err)
	}
	return u
}
func login(t *testing.T, s *account.Service) *account.Session {
	t.Helper()
	v, err := s.Login(context.Background(), "sample_member", password)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func TestRegistrationUniqueAndAuditRollback(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	out := make(chan error, 2)
	for _, s := range []*account.Service{f.a, f.b} {
		wg.Add(1)
		go func(s *account.Service) {
			defer wg.Done()
			_, err := s.Register(ctx, "sample_member", password, "会员")
			out <- err
		}(s)
	}
	wg.Wait()
	close(out)
	success, duplicate := 0, 0
	for err := range out {
		switch status.Code(err) {
		case codes.OK:
			success++
		case codes.AlreadyExists:
			duplicate++
		default:
			t.Fatal(err)
		}
	}
	if success != 1 || duplicate != 1 {
		t.Fatal(success, duplicate)
	}
	var count int64
	f.db.Model(&model.AuditEvent{}).Count(&count)
	if count != 1 {
		t.Fatal("non-atomic registration audit", count)
	}
	if err := f.db.Exec("CREATE TRIGGER fail_member_audit BEFORE INSERT ON audit_events FOR EACH ROW SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'fixture audit failure'").Error; err != nil {
		t.Fatal(err)
	}
	_, err := f.a.Register(ctx, "failed_member", password, "失败")
	code(t, err, codes.Unavailable)
	f.db.Model(&model.User{}).Where("username = ?", "failed_member").Count(&count)
	if count != 0 {
		t.Fatal("registration survived audit failure")
	}
}
func TestRefreshReplayAndConcurrentConsumption(t *testing.T) {
	f := setup(t)
	register(t, f)
	old := login(t, f.a)
	ctx := context.Background()
	var wg sync.WaitGroup
	out := make(chan *account.Session, 2)
	errs := make(chan error, 2)
	for _, s := range []*account.Service{f.a, f.b} {
		wg.Add(1)
		go func(s *account.Service) {
			defer wg.Done()
			v, err := s.Refresh(ctx, old.RefreshToken)
			out <- v
			errs <- err
		}(s)
	}
	wg.Wait()
	close(out)
	close(errs)
	ok, bad := 0, 0
	for err := range errs {
		if err == nil {
			ok++
		} else {
			code(t, err, codes.Unauthenticated)
			bad++
		}
	}
	if ok != 1 || bad != 1 {
		t.Fatal(ok, bad)
	}
	for v := range out {
		if v != nil {
			_, err := f.b.Current(ctx, v.AccessToken)
			code(t, err, codes.Unauthenticated)
		}
	}
	session := login(t, f.a)
	if err := f.b.Logout(ctx, session.RefreshToken); err != nil {
		t.Fatal(err)
	}
	_, err := f.a.Current(ctx, session.AccessToken)
	code(t, err, codes.Unauthenticated)
}
func TestProfileStatusAndPasswordRevocation(t *testing.T) {
	f := setup(t)
	u := register(t, f)
	session := login(t, f.a)
	ctx := context.Background()
	profile, err := f.b.UpdateProfile(ctx, session.AccessToken, "新昵称", account.Revision(u.UpdatedAt))
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.a.UpdateProfile(ctx, session.AccessToken, "覆盖", account.Revision(u.UpdatedAt))
	code(t, err, codes.Aborted)
	if err = f.manager.SetStatus(ctx, u.ID, false, account.Revision(profile.UpdatedAt)); err != nil {
		t.Fatal(err)
	}
	_, err = f.b.Current(ctx, session.AccessToken)
	code(t, err, codes.Unauthenticated)
	_, err = f.a.Refresh(ctx, session.RefreshToken)
	code(t, err, codes.Unauthenticated)
	user, events, err := f.manager.Get(ctx, u.ID)
	if err != nil || len(events) != 3 {
		t.Fatal(err, len(events))
	}
	if err = f.manager.SetStatus(ctx, u.ID, true, account.Revision(user.UpdatedAt)); err != nil {
		t.Fatal(err)
	}
	_, err = f.a.Current(ctx, session.AccessToken)
	code(t, err, codes.Unauthenticated)
	session = login(t, f.a)
	if err = f.b.ChangePassword(ctx, session.AccessToken, password, "member-test-password-two"); err != nil {
		t.Fatal(err)
	}
	_, err = f.a.Current(ctx, session.AccessToken)
	code(t, err, codes.Unauthenticated)
	_, err = f.a.Login(ctx, u.Username, password)
	code(t, err, codes.Unauthenticated)
	_, err = f.b.Login(ctx, u.Username, "member-test-password-two")
	if err != nil {
		t.Fatal(err)
	}
}
func TestProfileConcurrentAndAuthorizationFailClosed(t *testing.T) {
	f := setup(t)
	u := register(t, f)
	session := login(t, f.a)
	ctx := context.Background()
	var wg sync.WaitGroup
	out := make(chan error, 2)
	for i, s := range []*account.Service{f.a, f.b} {
		wg.Add(1)
		go func(i int, s *account.Service) {
			defer wg.Done()
			_, err := s.UpdateProfile(ctx, session.AccessToken, fmt.Sprintf("昵称%d", i), account.Revision(u.UpdatedAt))
			out <- err
		}(i, s)
	}
	wg.Wait()
	close(out)
	ok, conflict := 0, 0
	for err := range out {
		if err == nil {
			ok++
		} else {
			code(t, err, codes.Aborted)
			conflict++
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatal(ok, conflict)
	}
	for _, denial := range []codes.Code{codes.PermissionDenied, codes.Unauthenticated, codes.Unavailable} {
		manager := management.New(repository.NewAuth(f.db), checker{xerr.New(denial)})
		code(t, manager.SetStatus(ctx, u.ID, false, account.Revision(u.UpdatedAt)), denial)
		_, _, err := manager.List(ctx, management.Query{Page: 1, PageSize: 20, Enabled: "all"})
		code(t, err, denial)
	}
	current, err := f.a.Current(ctx, session.AccessToken)
	if err != nil || !current.Enabled {
		t.Fatal("unauthorized mutation", err)
	}
	if err = f.db.Exec("CREATE TRIGGER fail_member_audit BEFORE INSERT ON audit_events FOR EACH ROW SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'fixture audit failure'").Error; err != nil {
		t.Fatal(err)
	}
	code(t, f.manager.SetStatus(ctx, u.ID, false, account.Revision(current.UpdatedAt)), codes.Unavailable)
	_, err = f.b.Current(ctx, session.AccessToken)
	if err != nil {
		t.Fatal("revocation survived rollback", err)
	}
	_, err = f.b.UpdateProfile(ctx, session.AccessToken, "不应提交", account.Revision(current.UpdatedAt))
	code(t, err, codes.Unavailable)
	after, err := f.a.Current(ctx, session.AccessToken)
	if err != nil || after.Nickname != current.Nickname {
		t.Fatal("profile survived rollback", err)
	}
	code(t, f.b.ChangePassword(ctx, session.AccessToken, password, "member-test-password-two"), codes.Unavailable)
	_, err = f.a.Current(ctx, session.AccessToken)
	if err != nil {
		t.Fatal("password revocation survived rollback", err)
	}
}
