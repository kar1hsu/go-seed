//go:build integration

package auth

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/migrations"
	testidentity "github.com/kar1hsu/go-seed/tests/fixtures/identity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

const testPassword = "integration-password-one"

type fixture struct {
	db            *gorm.DB
	first, second *Service
	username      string
	id            uint64
}

func setup(t *testing.T) *fixture {
	t.Helper()
	port, err := strconv.Atoi(os.Getenv("MYSQL_PORT"))
	if err != nil {
		t.Fatal("run make test-auth with the local MySQL stack")
	}
	db, err := dbx.OpenMySQL(dbx.MySQLConfig{Host: os.Getenv("MYSQL_HOST"), Port: port, Database: os.Getenv("MYSQL_DATABASE"), User: os.Getenv("MYSQL_USER"), Password: os.Getenv("MYSQL_PASSWORD"), MaxOpenConns: 10, MaxIdleConns: 2, ConnMaxLifetime: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := db.DB()
	t.Cleanup(func() { pool.Close() })
	if err = migrations.Up(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
	if err = migrations.Up(context.Background(), pool); err != nil {
		t.Fatal("repeat migration", err)
	}
	tokens, err := identity.NewSigner(testidentity.Config(1, "admin", "test-admin", "test-api", 15*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	makeService := func() *Service {
		s, err := New(repository.NewAuth(db), tokens, 24*time.Hour, func(context.Context, string) error { return nil })
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	suffix, _ := identity.Random()
	username := "test_" + strings.ToLower(strings.NewReplacer("-", "a", "_", "b").Replace(suffix[:16]))
	f := &fixture{db: db, first: makeService(), second: makeService(), username: username}
	created, err := f.first.Bootstrap(context.Background(), username, testPassword, "Integration admin")
	if err != nil || !created {
		t.Fatal("bootstrap", created, err)
	}
	a, err := repository.NewAuth(db).ByUsername(context.Background(), username)
	if err != nil {
		t.Fatal(err)
	}
	f.id = a.ID
	t.Cleanup(func() {
		if err := db.Where("target_type = ? AND target_id = ? AND target_key = ?", "administrator", f.id, username).Delete(&model.AuditLog{}).Error; err != nil {
			t.Error(err)
		}
		if err := db.Where("id = ? AND username = ?", f.id, username).Delete(&model.Administrator{}).Error; err != nil {
			t.Error(err)
		}
	})
	return f
}
func login(t *testing.T, s *Service, name, password string) *Session {
	t.Helper()
	result, err := s.Login(context.Background(), name, password)
	if err != nil {
		t.Fatal("login", err)
	}
	return result
}
func wantCode(t *testing.T, err error, code codes.Code) {
	t.Helper()
	if status.Code(err) != code {
		t.Fatalf("got %v, expected %v", err, code)
	}
}

func TestBootstrapAndLogin(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	created, err := f.second.Bootstrap(ctx, f.username, "a-different-password", "Replacement")
	if err != nil || created {
		t.Fatal("bootstrap changed existing account")
	}
	session := login(t, f.first, f.username, testPassword)
	a, err := f.second.Current(ctx, session.AccessToken)
	if err != nil || a.DisplayName != "Integration admin" {
		t.Fatal("current", err)
	}
	if a.PasswordHash == testPassword || !strings.HasPrefix(a.PasswordHash, "$2a$") {
		t.Fatal("password not hashed")
	}
	_, err = f.first.Login(ctx, f.username, "wrong-password-value")
	wantCode(t, err, codes.Unauthenticated)
	_, err = f.first.Login(ctx, "missing_account", testPassword)
	wantCode(t, err, codes.Unauthenticated)
	_, err = f.first.Login(ctx, "INVALID", testPassword)
	wantCode(t, err, codes.InvalidArgument)
}
func TestRefreshReplayRevokesAcrossInstances(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	old := login(t, f.first, f.username, testPassword)
	current, err := f.second.Refresh(ctx, old.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if current.RefreshToken == old.RefreshToken {
		t.Fatal("refresh did not rotate")
	}
	if _, err = f.first.Current(ctx, current.AccessToken); err != nil {
		t.Fatal(err)
	}
	_, err = f.first.Refresh(ctx, old.RefreshToken)
	wantCode(t, err, codes.Unauthenticated)
	_, err = f.second.Current(ctx, current.AccessToken)
	wantCode(t, err, codes.Unauthenticated)
	_, err = f.second.Refresh(ctx, current.RefreshToken)
	wantCode(t, err, codes.Unauthenticated)
	var rows []model.AdminRefreshToken
	if err = f.db.Where("session_id = ?", strings.Split(current.RefreshToken, ".")[0]).Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if len(row.Digest) != 64 || row.Digest == current.RefreshToken {
			t.Fatal("raw refresh persisted")
		}
	}
}
func TestConcurrentRotation(t *testing.T) {
	f := setup(t)
	old := login(t, f.first, f.username, testPassword)
	var group sync.WaitGroup
	group.Add(2)
	start := make(chan struct{})
	results := make(chan *Session, 2)
	failures := make(chan error, 2)
	for _, service := range []*Service{f.first, f.second} {
		go func(s *Service) {
			defer group.Done()
			<-start
			result, err := s.Refresh(context.Background(), old.RefreshToken)
			if err != nil {
				failures <- err
			} else {
				results <- result
			}
		}(service)
	}
	close(start)
	group.Wait()
	close(results)
	close(failures)
	if len(results) != 1 || len(failures) != 1 {
		t.Fatalf("rotation winners=%d failures=%d", len(results), len(failures))
	}
	wantCode(t, <-failures, codes.Unauthenticated)
	winner := <-results
	_, err := f.first.Current(context.Background(), winner.AccessToken)
	wantCode(t, err, codes.Unauthenticated)
}
func TestLogoutAndForgedRefresh(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	session := login(t, f.first, f.username, testPassword)
	sid := strings.Split(session.RefreshToken, ".")[0]
	_, err := f.second.Refresh(ctx, sid+"."+strings.Repeat("x", 43))
	wantCode(t, err, codes.Unauthenticated)
	if _, err = f.first.Current(ctx, session.AccessToken); err != nil {
		t.Fatal("forged token revoked valid session")
	}
	if err = f.second.Logout(ctx, session.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if err = f.first.Logout(ctx, session.RefreshToken); err != nil {
		t.Fatal("logout not repeatable", err)
	}
	_, err = f.first.Current(ctx, session.AccessToken)
	wantCode(t, err, codes.Unauthenticated)
	_, err = f.first.Refresh(ctx, session.RefreshToken)
	wantCode(t, err, codes.Unauthenticated)
}
func TestPasswordAndDisableRevokeAllSessions(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	one := login(t, f.first, f.username, testPassword)
	two := login(t, f.second, f.username, testPassword)
	wantCode(t, f.first.ChangePassword(ctx, one.AccessToken, "wrong-password-value", "new-password-value"), codes.InvalidArgument)
	if _, err := f.first.Current(ctx, one.AccessToken); err != nil {
		t.Fatal("wrong password revoked session")
	}
	if err := f.first.ChangePassword(ctx, one.AccessToken, testPassword, "new-password-value"); err != nil {
		t.Fatal(err)
	}
	for _, session := range []*Session{one, two} {
		_, err := f.second.Current(ctx, session.AccessToken)
		wantCode(t, err, codes.Unauthenticated)
		_, err = f.second.Refresh(ctx, session.RefreshToken)
		wantCode(t, err, codes.Unauthenticated)
	}
	_, err := f.first.Login(ctx, f.username, testPassword)
	wantCode(t, err, codes.Unauthenticated)
	active := login(t, f.first, f.username, "new-password-value")
	if err = f.second.Disable(ctx, f.username); err != nil {
		t.Fatal(err)
	}
	_, err = f.first.Current(ctx, active.AccessToken)
	wantCode(t, err, codes.Unauthenticated)
	_, err = f.first.Refresh(ctx, active.RefreshToken)
	wantCode(t, err, codes.Unauthenticated)
	created, err := f.first.Bootstrap(ctx, f.username, testPassword, "Replacement")
	if err != nil || created {
		t.Fatal("bootstrap modified disabled account")
	}
	_, err = f.first.Login(ctx, f.username, "new-password-value")
	wantCode(t, err, codes.Unauthenticated)
	// Re-enabling an account must not revive its old sessions.
	if err = f.db.Model(&model.Administrator{}).Where("id = ?", f.id).Update("enabled", true).Error; err != nil {
		t.Fatal(err)
	}
	_, err = f.first.Current(ctx, active.AccessToken)
	wantCode(t, err, codes.Unauthenticated)
}
func TestSessionExpiry(t *testing.T) {
	f := setup(t)
	session := login(t, f.first, f.username, testPassword)
	sid := strings.Split(session.RefreshToken, ".")[0]
	if err := f.db.Model(&model.AdminSession{}).Where("id = ?", sid).Update("expires_at", time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	_, err := f.second.Current(context.Background(), session.AccessToken)
	wantCode(t, err, codes.Unauthenticated)
	_, err = f.second.Refresh(context.Background(), session.RefreshToken)
	wantCode(t, err, codes.Unauthenticated)
}
