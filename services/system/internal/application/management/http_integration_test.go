//go:build integration

package management

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/system/internal/application/auth"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/resources"
	"gorm.io/gorm"
)

func TestHTTPManagement(t *testing.T) {
	base := os.Getenv("RBAC_SMOKE_URL")
	if base == "" {
		t.Skip("run make rbac-smoke against the local admin API")
	}
	port, err := strconv.Atoi(os.Getenv("MYSQL_PORT"))
	require(t, err)
	db, err := dbx.OpenMySQL(dbx.MySQLConfig{Host: os.Getenv("MYSQL_HOST"), Port: port, Database: os.Getenv("MYSQL_DATABASE"), User: os.Getenv("MYSQL_USER"), Password: os.Getenv("MYSQL_PASSWORD"), MaxOpenConns: 5, MaxIdleConns: 2, ConnMaxLifetime: time.Minute})
	require(t, err)
	pool, _ := db.DB()
	t.Cleanup(func() { pool.Close() })
	tokens, err := identity.New(identity.Config{Secret: strings.Repeat("s", 48), Issuer: "fixture", Audience: "fixture", AccessTTL: time.Hour})
	require(t, err)
	as, err := auth.New(repository.NewAuth(db), tokens, time.Hour, func(_ context.Context, _ string) error { return nil })
	require(t, err)
	suffix := make([]byte, 6)
	_, err = rand.Read(suffix)
	require(t, err)
	prefix := "http_" + hex.EncodeToString(suffix)
	_, err = as.Bootstrap(ctx, prefix+"_root", password, "HTTP Test Root")
	require(t, err)
	var super model.Role
	require(t, db.Where("code = ? AND built_in = ?", resources.SuperRole, true).First(&super).Error)
	root, err := repository.NewAuth(db).ByUsername(ctx, prefix+"_root")
	require(t, err)
	store := repository.NewRBAC(db)
	// Cleanup is restricted to this run's random names, even after a failed assertion.
	t.Cleanup(func() {
		err := db.Transaction(func(cleanup *gorm.DB) error {
			tx := repository.NewRBAC(cleanup)
			if _, err := tx.LockPolicy(ctx); err != nil {
				return err
			}
			if err := cleanup.Where("actor_id = ? AND actor_username = ?", root.ID, prefix+"_root").Delete(&model.AuditLog{}).Error; err != nil {
				return err
			}
			if err := cleanup.Where("username IN ?", []string{prefix + "_root", prefix + "_reader"}).Delete(&model.Administrator{}).Error; err != nil {
				return err
			}
			if err := cleanup.Where("code = ?", prefix+"_role").Delete(&model.Role{}).Error; err != nil {
				return err
			}
			if err := cleanup.Where("code = ?", prefix+"_dict").Delete(&model.Dictionary{}).Error; err != nil {
				return err
			}
			if err := cleanup.Where("code = ?", prefix+"_setting").Delete(&model.Setting{}).Error; err != nil {
				return err
			}
			// Only this random fixture's jobs/files, after the HTTP worker checks have completed.
			if err := cleanup.Where("owner_id = ?", root.ID).Delete(&model.Task{}).Error; err != nil {
				return err
			}
			if err := cleanup.Where("owner_id = ? AND status = 'deleted'", root.ID).Delete(&model.File{}).Error; err != nil {
				return err
			}
			return tx.Bump(ctx)
		})
		if err != nil {
			t.Error(err)
		}
	})
	require(t, store.Transaction(ctx, true, func(tx *repository.RBAC) error {
		if err := tx.SetRoles(ctx, root.ID, []uint64{super.ID}, time.Now().UTC()); err != nil {
			return err
		}
		return tx.Bump(ctx)
	}))
	client := &http.Client{Timeout: 10 * time.Second}
	var lastTrace string
	call := func(method, path, token string, body any, csrf bool, want int) map[string]any {
		t.Helper()
		var payload []byte
		if body != nil {
			payload, err = json.Marshal(body)
			require(t, err)
		}
		req, err := http.NewRequest(method, base+"/admin-api/v1"+path, bytes.NewReader(payload))
		require(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", strconv.FormatUint(root.ID, 10))
		req.Header.Set("X-Roles", resources.SuperRole)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if csrf {
			req.Header.Set("X-GoSeed-CSRF", "1")
		}
		resp, err := client.Do(req)
		require(t, err)
		defer resp.Body.Close()
		var envelope struct {
			Code    int
			Data    map[string]any
			TraceID string `json:"traceId"`
		}
		require(t, json.NewDecoder(resp.Body).Decode(&envelope))
		if resp.StatusCode != want {
			t.Fatalf("%s %s: status %d want %d code %d", method, path, resp.StatusCode, want, envelope.Code)
		}
		if envelope.TraceID == "" {
			t.Fatal("missing trace ID")
		}
		lastTrace = envelope.TraceID
		return envelope.Data
	}
	login := func(username string) string {
		t.Helper()
		return call("POST", "/auth/login", "", map[string]string{"username": username, "password": password}, true, 200)["accessToken"].(string)
	}
	call("GET", "/system/administrators", "", nil, false, 401)
	rootToken := login(prefix + "_root")
	profile := call("GET", "/auth/me", rootToken, nil, false, 200)
	if len(profile["permissions"].([]any)) < 9 {
		t.Fatal("root catalog incomplete")
	}
	created := call("POST", "/system/administrators", rootToken, map[string]string{"username": prefix + "_reader", "displayName": "HTTP reader", "password": password}, true, 201)
	creationTrace := lastTrace
	userID, revision := created["id"].(string), created["revision"].(string)
	if _, ok := created["passwordHash"]; ok {
		t.Fatal("password hash exposed")
	}
	reader := login(prefix + "_reader")
	call("GET", "/system/administrators", reader, nil, false, 403)
	role := call("POST", "/system/roles", rootToken, map[string]string{"code": prefix + "_role", "name": "HTTP role", "description": "Smoke fixture"}, true, 201)
	rolePath := "/system/roles/" + role["id"].(string)
	grants := map[string]any{"permissions": []string{resources.UserRead}, "revision": role["revision"]}
	call("PUT", rolePath+"/permissions", rootToken, grants, false, 403)
	call("PUT", rolePath+"/permissions", rootToken, grants, true, 200)
	call("PUT", rolePath+"/permissions", rootToken, grants, true, 409)
	adminPath := "/system/administrators/" + userID
	call("PUT", adminPath+"/roles", rootToken, map[string]any{"roleIds": []string{role["id"].(string)}, "revision": revision}, true, 200)
	data := call("GET", "/system/administrators?page=1&pageSize=1&keyword="+prefix, reader, nil, false, 200)
	if data["total"].(float64) != 2 || len(data["items"].([]any)) != 1 {
		t.Fatal("filter/page mismatch")
	}
	call("GET", "/system/roles", reader, nil, false, 403)
	call("POST", "/system/roles", reader, map[string]string{"code": "forged_role", "name": "Forged", "description": ""}, true, 403)
	call("PATCH", "/system/administrators/"+strconv.FormatUint(root.ID, 10), reader, map[string]any{"displayName": "Forged", "enabled": false, "revision": "forged"}, true, 403)
	call("GET", "/system/administrators?page=0", rootToken, nil, false, 400)
	var currentRole model.Role
	require(t, db.First(&currentRole, role["id"]).Error)
	call("DELETE", rolePath+"?revision="+url.QueryEscape(Revision(currentRole.UpdatedAt)), rootToken, nil, true, 409)
	var currentAdmin model.Administrator
	require(t, db.First(&currentAdmin, userID).Error)
	call("PUT", adminPath+"/roles", rootToken, map[string]any{"roleIds": []string{}, "revision": Revision(currentAdmin.UpdatedAt)}, true, 200)
	call("GET", "/system/administrators", reader, nil, false, 403)
	call("DELETE", rolePath+"?revision="+url.QueryEscape(Revision(currentRole.UpdatedAt)), rootToken, nil, true, 200)
	call("DELETE", rolePath+"?revision="+url.QueryEscape(Revision(currentRole.UpdatedAt)), rootToken, nil, true, 404)
	// Audit reads require their own permission, even if caller-supplied identity headers claim root.
	call("GET", "/system/audit-logs", "", nil, false, 401)
	call("GET", "/system/audit-logs", reader, nil, false, 403)
	logs := call("GET", "/system/audit-logs?actorId="+strconv.FormatUint(root.ID, 10), rootToken, nil, false, 200)
	if logs["total"].(float64) != 6 || logs["page"].(float64) != 1 || logs["pageSize"].(float64) != 20 {
		t.Fatal("audit count/default pagination mismatch")
	}
	logID := logs["items"].([]any)[0].(map[string]any)["id"].(string)
	call("GET", "/system/audit-logs/"+logID, reader, nil, false, 403)
	detail := call("GET", "/system/audit-logs/"+logID, rootToken, nil, false, 200)
	if detail["record"].(map[string]any)["action"] != resources.RoleDelete || len(detail["changes"].([]any)) == 0 {
		t.Fatal("deleted role audit missing")
	}
	filtered := call("GET", "/system/audit-logs?action="+resources.UserCreate+"&traceId="+creationTrace, rootToken, nil, false, 200)
	if filtered["total"].(float64) != 1 {
		t.Fatal("HTTP-to-RPC audit trace not propagated")
	}
	for _, query := range []string{"page=0", "pageSize=101", "action=forged", "actorId=01", "traceId=bad", "from=bad"} {
		call("GET", "/system/audit-logs?"+query, rootToken, nil, false, 400)
	}
	call("GET", "/system/audit-logs/0", rootToken, nil, false, 400)
	call("GET", "/system/audit-logs/18446744073709551615", rootToken, nil, false, 404)
	verifySettingsHTTP(t, call, db, prefix, rootToken, reader)
	verifyFilesTasksHTTP(t, call, db, prefix, rootToken, reader)
}
