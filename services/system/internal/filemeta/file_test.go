package filemeta

import (
	"context"
	"github.com/kar1hsu/go-seed/pkg/storage"
	"github.com/kar1hsu/go-seed/pkg/timezone"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"strings"
	"testing"
	"time"
)

func TestContentIsVerified(t *testing.T) {
	s, err := storage.New(storage.Config{Driver: "local", Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		data, kind string
		valid      bool
	}{{"hello", "text/plain", true}, {"<script>alert(1)</script>", "image/png", false}, {"{bad json}", "application/json", false}, {`{"ok":true}`, "application/json", true}} {
		key := Key(time.Now(), time.UTC)
		if err = s.Put(context.Background(), key, strings.NewReader(c.data), int64(len(c.data))); err != nil {
			t.Fatal(err)
		}
		_, err = Inspect(context.Background(), s, model.File{ObjectKey: key, ContentType: c.kind, Size: int64(len(c.data))})
		if (err == nil) != c.valid {
			t.Fatalf("%s: %v", c.kind, err)
		}
	}
	if Valid("../x", "text/plain", 5) || Valid("x.svg", "image/svg+xml", 10) || Valid("x", "text/plain", 0) {
		t.Fatal("unsafe metadata accepted")
	}
}

func TestKeyUsesUTCDateAndRandomName(t *testing.T) {
	created := time.Date(2026, 11, 1, 1, 30, 0, 0, time.FixedZone("UTC+8", 8*3600))
	first, second := Key(created, time.UTC), Key(created, time.UTC)
	if !strings.HasPrefix(first, "20261031/") || len(first) != 41 || first == second {
		t.Fatalf("unexpected dated keys: %q %q", first, second)
	}
	for _, c := range first[9:] {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			t.Fatalf("invalid random object name: %q", first)
		}
	}
}

func TestKeyUsesConfiguredCalendarAcrossMidnightAndDST(t *testing.T) {
	for _, tc := range []struct{ zone, instant, date string }{
		{"Asia/Shanghai", "2026-10-30T16:05:00Z", "20261031"},
		{"UTC", "2026-10-30T16:05:00Z", "20261030"},
		{"America/New_York", "2026-01-01T00:05:00Z", "20251231"},
		{"America/New_York", "2026-11-01T03:30:00Z", "20261031"},
		{"America/New_York", "2026-11-01T04:30:00Z", "20261101"},
		{"America/New_York", "2026-11-01T06:30:00Z", "20261101"},
	} {
		t.Run(tc.zone+"/"+tc.instant, func(t *testing.T) {
			location, err := timezone.Load(tc.zone)
			if err != nil {
				t.Fatal(err)
			}
			instant, err := time.Parse(time.RFC3339, tc.instant)
			if err != nil {
				t.Fatal(err)
			}
			if key := Key(instant, location); !strings.HasPrefix(key, tc.date+"/") || len(key) != 41 {
				t.Fatalf("wrong calendar prefix: %s", key)
			}
		})
	}
}
