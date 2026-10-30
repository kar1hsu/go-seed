package clientip

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrustBoundary(t *testing.T) {
	for _, tc := range []struct{ peer, xff, want string }{
		{"198.51.100.1:80", "1.2.3.4", "198.51.100.1"},
		{"10.0.0.2:80", "1.2.3.4, 198.51.100.1", "198.51.100.1"},
		{"10.0.0.2:80", "198.51.100.1, 10.0.0.3", "198.51.100.1"},
		{"10.0.0.2:80", "garbage", "10.0.0.2"},
		{"[::ffff:198.51.100.1]:80", "1.2.3.4", "198.51.100.1"},
	} {
		t.Run(tc.peer+tc.xff, func(t *testing.T) {
			m, err := Middleware("10.0.0.0/24")
			if err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tc.peer
			r.Header.Set("X-Forwarded-For", tc.xff)
			m(func(w http.ResponseWriter, r *http.Request) {
				if got := From(r.Context()); got != tc.want {
					t.Fatalf("got %s want %s", got, tc.want)
				}
			})(httptest.NewRecorder(), r)
		})
	}
	if _, err := Middleware("all"); err == nil {
		t.Fatal("invalid trust configuration accepted")
	}
}
