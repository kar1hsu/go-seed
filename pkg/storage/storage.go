// Package storage provides private, immutable object storage. It owns no business metadata.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var ErrExists = errors.New("object already exists")
var ErrSize = errors.New("object size mismatch")

const MaxSize int64 = 10 << 20

type Config struct {
	Driver    string `json:",default=local"`
	Root      string `json:",default=.data/files"`
	Endpoint  string `json:",optional"`
	Bucket    string `json:",optional"`
	AccessKey string `json:",optional"`
	SecretKey string `json:",optional"`
	Secure    bool   `json:",default=true"`
	Region    string `json:",default=us-east-1"`
}
type Store interface {
	Put(context.Context, string, io.Reader, int64) error
	Open(context.Context, string) (io.ReadCloser, error)
	Delete(context.Context, string) error
	Check(context.Context) error
	CleanupTemporary(context.Context, time.Time) error
}

func validKey(key string) bool {
	// Flat keys remain readable for files created before dated directories.
	if len(key) == 41 {
		if key[8] != '/' {
			return false
		}
		date, err := time.Parse("20060102", key[:8])
		if err != nil || date.Year() < 1 || date.Format("20060102") != key[:8] {
			return false
		}
		key = key[9:]
	}
	if len(key) != 32 {
		return false
	}
	for _, c := range key {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func New(c Config) (Store, error) {
	switch c.Driver {
	case "local":
		if c.Root == "" {
			return nil, errors.New("missing storage root")
		}
		root, err := filepath.Abs(c.Root)
		if err != nil {
			return nil, err
		}
		if err = os.MkdirAll(root, 0700); err != nil {
			return nil, err
		}
		return &local{root}, nil
	case "s3":
		if c.Endpoint == "" || c.Bucket == "" || c.AccessKey == "" || c.SecretKey == "" {
			return nil, errors.New("incomplete S3 configuration")
		}
		client, err := minio.New(c.Endpoint, &minio.Options{Creds: credentials.NewStaticV4(c.AccessKey, c.SecretKey, ""), Secure: c.Secure, Region: c.Region})
		if err != nil {
			return nil, err
		}
		tempDir := filepath.Join(c.Root, ".spool")
		if err := os.MkdirAll(tempDir, 0700); err != nil {
			return nil, err
		}
		return &s3{client: client, bucket: c.Bucket, tempDir: tempDir}, nil
	default:
		return nil, fmt.Errorf("unsupported storage driver %q", c.Driver)
	}
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

// Spooling enforces exact size and keeps incomplete objects invisible; memory stays bounded.
func spool(ctx context.Context, dir string, r io.Reader, size int64) (*os.File, error) {
	if size < 1 || size > MaxSize {
		return nil, ErrSize
	}
	f, err := os.CreateTemp(dir, ".goseed-upload-*")
	if err != nil {
		return nil, err
	}
	ok := false
	defer func() {
		if !ok {
			f.Close()
			os.Remove(f.Name())
		}
	}()
	n, err := io.Copy(f, io.LimitReader(contextReader{ctx, r}, size+1))
	if err != nil {
		return nil, err
	}
	if n != size {
		return nil, ErrSize
	}
	if err = f.Sync(); err != nil {
		return nil, err
	}
	if _, err = f.Seek(0, 0); err != nil {
		return nil, err
	}
	ok = true
	return f, nil
}

type local struct{ root string }

func (s *local) Put(ctx context.Context, key string, r io.Reader, size int64) error {
	if !validKey(key) {
		return errors.New("invalid object key")
	}
	f, err := spool(ctx, s.root, r, size)
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if err = ctx.Err(); err != nil {
		return err
	}
	root, err := os.OpenRoot(s.root)
	if err != nil {
		return err
	}
	defer root.Close()
	directory := path.Dir(key)
	if err = root.MkdirAll(directory, 0700); err != nil {
		return err
	}
	// Root confines every directory operation, including symlink traversal.
	err = root.Link(filepath.Base(f.Name()), key)
	if os.IsExist(err) {
		return ErrExists
	}
	if err != nil {
		return err
	}
	// Persist the object entry and any newly created date directories.
	for {
		dir, err := root.Open(directory)
		if err != nil {
			return err
		}
		err = dir.Sync()
		dir.Close()
		if err != nil || directory == "." {
			return err
		}
		directory = path.Dir(directory)
	}
}
func (s *local) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if !validKey(key) {
		return nil, errors.New("invalid object key")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(s.root)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return root.Open(key)
}
func (s *local) Delete(ctx context.Context, key string) error {
	if !validKey(key) {
		return errors.New("invalid object key")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	root, err := os.OpenRoot(s.root)
	if err != nil {
		return err
	}
	defer root.Close()
	err = root.Remove(key)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
func (s *local) Check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := os.Stat(s.root)
	return err
}

type s3 struct {
	client  *minio.Client
	bucket  string
	tempDir string
}

func (s *s3) Put(ctx context.Context, key string, r io.Reader, size int64) error {
	if !validKey(key) {
		return errors.New("invalid object key")
	}
	f, err := spool(ctx, s.tempDir, r, size)
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	opts := minio.PutObjectOptions{ContentType: "application/octet-stream", DisableMultipart: true}
	opts.SetMatchETagExcept("*")
	_, err = s.client.PutObject(ctx, s.bucket, key, f, size, opts)
	if minio.ToErrorResponse(err).StatusCode == 412 {
		return ErrExists
	}
	return err
}
func (s *s3) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if !validKey(key) {
		return nil, errors.New("invalid object key")
	}
	o, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	if _, err = o.Stat(); err != nil {
		o.Close()
		return nil, err
	}
	return o, nil
}
func (s *s3) Delete(ctx context.Context, key string) error {
	if !validKey(key) {
		return errors.New("invalid object key")
	}
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}
func (s *s3) Check(ctx context.Context) error {
	ok, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("storage bucket missing")
	}
	return nil
}

func cleanupTemporary(ctx context.Context, dir string, before time.Time) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err = ctx.Err(); err != nil {
			return err
		}
		if !strings.HasPrefix(entry.Name(), ".goseed-upload-") || !entry.Type().IsRegular() {
			continue
		}
		info, err := entry.Info()
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if info.ModTime().Before(before) {
			if err = os.Remove(filepath.Join(dir, entry.Name())); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return nil
}
func (s *local) CleanupTemporary(ctx context.Context, before time.Time) error {
	return cleanupTemporary(ctx, s.root, before)
}
func (s *s3) CleanupTemporary(ctx context.Context, before time.Time) error {
	return cleanupTemporary(ctx, s.tempDir, before)
}
