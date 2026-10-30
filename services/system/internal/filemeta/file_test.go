package filemeta

import (
	"context"
	"github.com/kar1hsu/go-seed/pkg/storage"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"strings"
	"testing"
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
		key := Key()
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
