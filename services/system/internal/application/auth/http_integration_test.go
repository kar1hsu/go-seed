//go:build integration

package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/kar1hsu/go-seed/pkg/identity"
)

type httpEnvelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		Admin        struct {
			Username string `json:"username"`
		} `json:"admin"`
	} `json:"data"`
	TraceID string `json:"traceId"`
}

func TestHTTPAuthLifecycle(t *testing.T) {
	base := os.Getenv("AUTH_SMOKE_URL")
	if base == "" {
		t.Skip("set AUTH_SMOKE_URL to run against a live admin API")
	}
	f := setup(t)
	client := &http.Client{Timeout: 10 * time.Second}
	request := func(method, path string, body any, cookie, access, origin string, csrf bool, want int) (httpEnvelope, string) {
		t.Helper()
		payload, _ := json.Marshal(body)
		req, err := http.NewRequest(method, base+"/admin-api/v1/auth"+path, bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "1") // Untrusted identity headers must not grant access.
		if csrf {
			req.Header.Set("X-GoSeed-CSRF", "1")
		}
		if cookie != "" {
			req.AddCookie(&http.Cookie{Name: "goseed_refresh", Value: cookie})
		}
		if access != "" {
			req.Header.Set("Authorization", "Bearer "+access)
		}
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		var envelope httpEnvelope
		if err = json.Unmarshal(raw, &envelope); err != nil {
			t.Fatalf("%s %s: invalid JSON status %d", method, path, resp.StatusCode)
		}
		if resp.StatusCode != want {
			t.Fatalf("%s %s: status %d expected %d code=%d", method, path, resp.StatusCode, want, envelope.Code)
		}
		if resp.Header.Get("Cache-Control") != "no-store" {
			t.Fatal("auth response can be cached")
		}
		if envelope.Data.RefreshToken != "" {
			t.Fatal("refresh token leaked into JSON")
		}
		if envelope.TraceID == "" {
			t.Fatal("missing trace ID")
		}
		nextCookie := ""
		for _, item := range resp.Cookies() {
			if item.Name == "goseed_refresh" {
				if !item.HttpOnly || item.Path != "/admin-api/v1/auth" || item.SameSite != http.SameSiteStrictMode || item.Domain != "" {
					t.Fatal("invalid cookie policy")
				}
				nextCookie = item.Value
			}
		}
		return envelope, nextCookie
	}
	credentials := map[string]string{"username": f.username, "password": testPassword}
	request("POST", "/login", credentials, "", "", "", false, 403)
	request("POST", "/login", credentials, "", "", "https://foreign.example", true, 403)
	request("GET", "/me", nil, "", "", "", false, 401)
	logged, cookie := request("POST", "/login", credentials, "", "", base, true, 200)
	if cookie == "" || logged.Data.AccessToken == "" || logged.Data.Admin.Username != f.username {
		t.Fatal("missing login result")
	}
	request("GET", "/me", nil, "", logged.Data.AccessToken, "", false, 200)
	var claims identity.Claims
	_, err := jwt.ParseWithClaims(logged.Data.AccessToken, &claims, func(*jwt.Token) (any, error) { return []byte(os.Getenv("ADMIN_TOKEN_SECRET")), nil })
	if err != nil {
		t.Fatal("cannot inspect issued claims", err)
	}
	claims.Kind = "member"
	memberToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(os.Getenv("ADMIN_TOKEN_SECRET")))
	if err != nil {
		t.Fatal(err)
	}
	request("GET", "/me", nil, "", memberToken, "", false, 401)

	refreshed, next := request("POST", "/refresh", nil, cookie, "", base, true, 200)
	if next == cookie || next == "" {
		t.Fatal("cookie not rotated")
	}
	request("POST", "/refresh", nil, cookie, "", base, true, 401)
	request("GET", "/me", nil, "", refreshed.Data.AccessToken, "", false, 401)
	request("POST", "/refresh", nil, next, "", base, true, 401)
	logged, cookie = request("POST", "/login", credentials, "", "", base, true, 200)
	request("POST", "/password", map[string]string{"currentPassword": "incorrect-password", "newPassword": "http-test-new-password"}, cookie, logged.Data.AccessToken, base, true, 400)
	request("GET", "/me", nil, "", logged.Data.AccessToken, "", false, 200)
	_, cleared := request("POST", "/password", map[string]string{"currentPassword": testPassword, "newPassword": "http-test-new-password"}, cookie, logged.Data.AccessToken, base, true, 200)
	if cleared != "" {
		t.Fatal("cookie not cleared after password change")
	}
	request("GET", "/me", nil, "", logged.Data.AccessToken, "", false, 401)
	request("POST", "/refresh", nil, cookie, "", base, true, 401)
	request("POST", "/login", credentials, "", "", base, true, 401)
	credentials["password"] = "http-test-new-password"
	logged, cookie = request("POST", "/login", credentials, "", "", base, true, 200)
	request("POST", "/logout", nil, cookie, "", base, true, 200)
	request("GET", "/me", nil, "", logged.Data.AccessToken, "", false, 401)
	request("POST", "/logout", nil, "", "", base, true, 200)
}
