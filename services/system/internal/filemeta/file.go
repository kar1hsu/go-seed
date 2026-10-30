package filemeta

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/kar1hsu/go-seed/pkg/storage"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"google.golang.org/grpc/codes"
	"io"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"
)

func Key() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func Valid(name, kind string, size int64) bool {
	if !utf8.ValidString(name) || len([]rune(name)) > 180 || strings.TrimSpace(name) == "" || strings.ContainsAny(name, "/\\") || strings.ContainsFunc(name, unicode.IsControl) || size < 1 || size > storage.MaxSize {
		return false
	}
	switch kind {
	case "image/png", "image/jpeg", "application/pdf", "text/plain", "text/csv", "application/json", "application/zip":
		return true
	}
	return false
}
func Inspect(ctx context.Context, objects storage.Store, f model.File) (string, error) {
	r, err := objects.Open(ctx, f.ObjectKey)
	if err != nil {
		return "", err
	}
	defer r.Close()
	// File size is bounded to 10 MiB; sniffing and hashing use the actual bytes, never browser metadata.
	h := sha256.New()
	head := make([]byte, 512)
	n, err := io.ReadFull(r, head)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", err
	}
	head = head[:n]
	h.Write(head)
	size, err := io.Copy(h, io.LimitReader(r, storage.MaxSize+1-int64(n)))
	if err != nil {
		return "", err
	}
	if size+int64(n) != f.Size {
		return "", xerr.New(codes.InvalidArgument)
	}
	detected := strings.Split(http.DetectContentType(head), ";")[0]
	want := f.ContentType
	if want == "text/csv" || want == "application/json" {
		want = "text/plain"
	}
	if detected != want {
		return "", xerr.New(codes.InvalidArgument)
	}
	if f.ContentType == "application/json" {
		other, err := objects.Open(ctx, f.ObjectKey)
		if err != nil {
			return "", err
		}
		defer other.Close()
		data, err := io.ReadAll(io.LimitReader(other, storage.MaxSize+1))
		if err != nil {
			return "", err
		}
		if !json.Valid(data) {
			return "", xerr.New(codes.InvalidArgument)
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
