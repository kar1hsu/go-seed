//go:build integration

package account_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/services/member/internal/model"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type response struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Trace   string          `json:"traceId"`
}

func TestHTTPMember(t *testing.T) {
	if os.Getenv("MEMBER_SMOKE") != "1" {
		t.Skip("run make member-smoke against the running stack")
	}
	app := "http://127.0.0.1:" + os.Getenv("APP_API_PORT")
	admin := "http://127.0.0.1:" + os.Getenv("ADMIN_API_PORT")
	client := &http.Client{Timeout: 8 * time.Second}
	call := func(base, method, path string, body any, token, cookie string, csrf bool, want int) (response, *http.Response) {
		t.Helper()
		var reader io.Reader
		if body != nil {
			data, _ := json.Marshal(body)
			reader = bytes.NewReader(data)
		}
		req, err := http.NewRequest(method, base+path, reader)
		if err != nil {
			t.Fatal(err)
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}
		if csrf {
			req.Header.Set("X-GoSeed-CSRF", "1")
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var result response
		if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatal(path, "invalid JSON response")
		}
		if resp.StatusCode != want {
			t.Fatalf("%s %s status=%d expected=%d code=%d trace=%s", method, path, resp.StatusCode, want, result.Code, result.Trace)
		}
		if result.Trace == "" {
			t.Fatal("missing trace")
		}
		return result, resp
	}
	decode := func(r response) map[string]any {
		t.Helper()
		var data map[string]any
		if err := json.Unmarshal(r.Data, &data); err != nil {
			t.Fatal(err)
		}
		return data
	}
	cookieValue := func(r *http.Response, name string) string {
		t.Helper()
		for _, c := range r.Cookies() {
			if c.Name == name {
				if !c.HttpOnly || c.SameSite != http.SameSiteStrictMode {
					t.Fatal("cookie security attributes")
				}
				return name + "=" + c.Value
			}
		}
		t.Fatal("missing cookie", name)
		return ""
	}
	suffix, _ := identity.Random()
	username := "smoke_" + strings.ToLower(strings.NewReplacer("-", "a", "_", "b").Replace(suffix[:12]))
	// Cleanup only this unique fixture; the application intentionally has no member deletion endpoint.
	port, _ := strconv.Atoi(os.Getenv("MYSQL_PORT"))
	db, err := dbx.OpenMySQL(dbx.MySQLConfig{Host: os.Getenv("MYSQL_HOST"), Port: port, Database: os.Getenv("MEMBER_MYSQL_DATABASE"), User: os.Getenv("MEMBER_MYSQL_USER"), Password: os.Getenv("MEMBER_MYSQL_PASSWORD"), MaxOpenConns: 2, MaxIdleConns: 1, ConnMaxLifetime: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := db.DB()
	t.Cleanup(func() {
		defer pool.Close()
		if err := db.Where("username = ?", username).Delete(&model.User{}).Error; err != nil {
			t.Error("member fixture cleanup failed")
		}
	})
	input := map[string]any{"username": username, "password": password, "nickname": "接口验收会员"}
	call(app, "POST", "/api/v1/auth/register", input, "", "", false, 403)
	registered, _ := call(app, "POST", "/api/v1/auth/register", input, "", "", true, 201)
	user := decode(registered)
	id := user["id"].(string)
	call(app, "POST", "/api/v1/auth/register", input, "", "", true, 409)
	for _, key := range []string{"passwordHash", "enabled", "tokenVersion", "refreshToken"} {
		if _, ok := user[key]; ok {
			t.Fatal("private field in member profile", key)
		}
	}
	loginInput := map[string]string{"username": username, "password": password}
	auth, headers := call(app, "POST", "/api/v1/auth/login", loginInput, "", "", true, 200)
	session := decode(auth)
	token := session["accessToken"].(string)
	cookie := cookieValue(headers, "goseed_member_refresh")
	if bytes.Contains(auth.Data, []byte("refreshToken")) {
		t.Fatal("refresh token exposed in JSON")
	}
	for _, c := range headers.Cookies() {
		if c.Name == "goseed_member_refresh" && c.Path != "/api/v1/auth" {
			t.Fatal("wrong cookie path")
		}
	}
	call(admin, "GET", "/admin-api/v1/auth/me", nil, token, "", false, 401)
	current, _ := call(app, "GET", "/api/v1/me", nil, token, "", false, 200)
	profile := decode(current)
	updated, _ := call(app, "PATCH", "/api/v1/me", map[string]any{"nickname": "更新后的昵称", "revision": profile["revision"], "id": "999999"}, token, "", true, 200)
	if decode(updated)["id"] != id {
		t.Fatal("client supplied ID changed ownership")
	}
	call(app, "PATCH", "/api/v1/me", map[string]any{"nickname": "旧版本", "revision": profile["revision"]}, token, "", true, 409)
	rotated, rotatedHeaders := call(app, "POST", "/api/v1/auth/refresh", nil, "", cookie, true, 200)
	newToken := decode(rotated)["accessToken"].(string)
	newCookie := cookieValue(rotatedHeaders, "goseed_member_refresh")
	call(app, "POST", "/api/v1/auth/refresh", nil, "", cookie, true, 401)
	call(app, "GET", "/api/v1/me", nil, newToken, "", false, 401)
	auth, headers = call(app, "POST", "/api/v1/auth/login", loginInput, "", "", true, 200)
	token = decode(auth)["accessToken"].(string)
	cookie = cookieValue(headers, "goseed_member_refresh")
	rootLogin, rootHeaders := call(admin, "POST", "/admin-api/v1/auth/login", map[string]string{"username": "admin", "password": os.Getenv("ADMIN_BOOTSTRAP_PASSWORD")}, "", "", true, 200)
	rootToken := decode(rootLogin)["accessToken"].(string)
	rootCookie := cookieValue(rootHeaders, "goseed_refresh")
	defer call(admin, "POST", "/admin-api/v1/auth/logout", nil, "", rootCookie, true, 200)
	call(app, "GET", "/api/v1/me", nil, rootToken, "", false, 401)
	call(app, "POST", "/api/v1/auth/refresh", nil, "", rootCookie, true, 401)
	call(admin, "POST", "/admin-api/v1/auth/refresh", nil, "", newCookie, true, 401)
	call(admin, "GET", "/admin-api/v1/member/users", nil, token, "", false, 401)
	list, _ := call(admin, "GET", "/admin-api/v1/member/users?keyword="+username, nil, rootToken, "", false, 200)
	if decode(list)["total"].(float64) != 1 {
		t.Fatal("member listing")
	}
	detail, _ := call(admin, "GET", "/admin-api/v1/member/users/"+id, nil, rootToken, "", false, 200)
	record := decode(detail)["user"].(map[string]any)
	call(admin, "PUT", "/admin-api/v1/member/users/"+id+"/status", map[string]any{"enabled": false, "revision": record["revision"]}, rootToken, "", true, 200)
	call(app, "GET", "/api/v1/me", nil, token, "", false, 401)
	call(app, "POST", "/api/v1/auth/refresh", nil, "", cookie, true, 401)
	call(app, "POST", "/api/v1/auth/login", loginInput, "", "", true, 401)
	call(admin, "PUT", "/admin-api/v1/member/users/"+id+"/status", map[string]any{"enabled": true, "revision": record["revision"]}, rootToken, "", true, 409)
	detail, _ = call(admin, "GET", "/admin-api/v1/member/users/"+id, nil, rootToken, "", false, 200)
	record = decode(detail)["user"].(map[string]any)
	call(admin, "PUT", "/admin-api/v1/member/users/"+id+"/status", map[string]any{"enabled": true, "revision": record["revision"]}, rootToken, "", true, 200)
	call(app, "GET", "/api/v1/me", nil, token, "", false, 401)
	auth, headers = call(app, "POST", "/api/v1/auth/login", loginInput, "", "", true, 200)
	token = decode(auth)["accessToken"].(string)
	cookie = cookieValue(headers, "goseed_member_refresh")
	call(app, "POST", "/api/v1/me/password", map[string]string{"currentPassword": password, "newPassword": "member-http-password-two"}, token, "", true, 200)
	call(app, "GET", "/api/v1/me", nil, token, "", false, 401)
	call(app, "POST", "/api/v1/auth/refresh", nil, "", cookie, true, 401)
	call(app, "POST", "/api/v1/auth/login", loginInput, "", "", true, 401)
	loginInput["password"] = "member-http-password-two"
	auth, headers = call(app, "POST", "/api/v1/auth/login", loginInput, "", "", true, 200)
	token = decode(auth)["accessToken"].(string)
	cookie = cookieValue(headers, "goseed_member_refresh")
	call(app, "POST", "/api/v1/auth/logout", nil, "", cookie, true, 200)
	call(app, "GET", "/api/v1/me", nil, token, "", false, 401)
	// A live administrator with no member grants must be rejected at the domain boundary.
	sys, err := dbx.OpenMySQL(dbx.MySQLConfig{Host: os.Getenv("MYSQL_HOST"), Port: port, Database: os.Getenv("MYSQL_DATABASE"), User: os.Getenv("MYSQL_USER"), Password: os.Getenv("MYSQL_PASSWORD"), MaxOpenConns: 2, MaxIdleConns: 1, ConnMaxLifetime: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	sysPool, _ := sys.DB()
	t.Cleanup(func() {
		defer sysPool.Close()
		sys.Exec("DELETE FROM audit_logs WHERE target_type = 'administrator' AND target_key = ?", username)
		sys.Exec("DELETE FROM administrators WHERE username = ?", username)
	})
	call(admin, "POST", "/admin-api/v1/system/administrators", map[string]string{"username": username, "password": password, "displayName": "会员权限验收"}, rootToken, "", true, 201)
	limited, _ := call(admin, "POST", "/admin-api/v1/auth/login", map[string]string{"username": username, "password": password}, "", "", true, 200)
	limitedToken := decode(limited)["accessToken"].(string)
	call(admin, "GET", "/admin-api/v1/member/users", nil, limitedToken, "", false, 403)
	call(admin, "PUT", "/admin-api/v1/member/users/"+id+"/status", map[string]any{"enabled": false, "revision": record["revision"]}, limitedToken, "", true, 403)
	// Dedicated credentials must not cross-read another domain's database.
	var ignored int
	if err = db.WithContext(context.Background()).Raw("SELECT 1 FROM `" + os.Getenv("MYSQL_DATABASE") + "`.administrators LIMIT 1").Scan(&ignored).Error; err == nil {
		t.Fatal("member DB account can read system tables")
	}
	if err = sys.Raw("SELECT 1 FROM `" + os.Getenv("MEMBER_MYSQL_DATABASE") + "`.users LIMIT 1").Scan(&ignored).Error; err == nil {
		t.Fatal("system DB account can read member tables")
	}
}
