package authhttp

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestRefreshCookieAttributes(t *testing.T) {
	recorder := httptest.NewRecorder()
	SetCookie(recorder, "opaque", time.Now().Add(time.Hour), true)
	cookie := recorder.Result().Cookies()[0]
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != 3 || cookie.Path != "/admin-api/v1/auth" || cookie.Domain != "" || cookie.MaxAge <= 0 {
		t.Fatal("cookie attributes invalid")
	}
	recorder = httptest.NewRecorder()
	SetCookie(recorder, "", time.Time{}, true)
	if recorder.Result().Cookies()[0].MaxAge != -1 {
		t.Fatal("cookie not expired")
	}
	request := httptest.NewRequest("POST", "/admin-api/v1/auth/refresh", nil)
	request.Header.Set("Cookie", "goseed_refresh=one; goseed_refresh=two")
	if _, err := ReadCookie(request); err == nil {
		t.Fatal("ambiguous cookie accepted")
	}
}
