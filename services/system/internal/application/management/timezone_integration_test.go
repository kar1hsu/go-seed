//go:build integration

package management

import (
	"strings"
	"testing"
	"time"

	"github.com/kar1hsu/go-seed/pkg/timezone"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
)

func TestFileCalendarUsesConfiguredTimezone(t *testing.T) {
	f := setup(t)
	// Pick an IANA zone whose calendar date differs from UTC at this instant,
	// so this test detects accidental UTC conversion at any time of day.
	zone := "Pacific/Kiritimati"
	if time.Now().UTC().Hour() < 12 {
		zone = "Etc/GMT+12"
	}
	location, err := timezone.Load(zone)
	require(t, err)
	service, err := New(repository.NewRBAC(f.db), f.a.authorizer, nil, location)
	require(t, err)
	check := func(file model.File) {
		t.Helper()
		prefix := file.CreatedAt.In(location).Format("20060102/")
		if prefix == file.CreatedAt.UTC().Format("20060102/") || !strings.HasPrefix(file.ObjectKey, prefix) {
			t.Fatalf("configured date was ignored: zone=%s key=%s created=%s", zone, file.ObjectKey, file.CreatedAt)
		}
	}
	upload, err := service.CreateFile(ctx, f.root, "notes.txt", "text/plain", 5)
	require(t, err)
	check(*upload)
	job, err := service.ExportDictionaries(ctx, f.root, "6fd751c2-44ab-4c6d-8d04-80b2e4e0dc59")
	require(t, err)
	var file model.File
	require(t, f.db.First(&file, job.FileID).Error)
	check(file)
	if file.Name != "dictionaries-"+file.CreatedAt.In(location).Format("20060102-150405")+".csv" {
		t.Fatal("export filename uses a different timezone", file.Name)
	}
	oldKey := file.ObjectKey
	require(t, f.db.Model(&file).Update("status", "deleted").Error)
	require(t, f.db.Model(job).Update("status", "failed").Error)
	require(t, service.RetryTask(ctx, f.root, job.ID))
	require(t, f.db.First(&file, job.FileID).Error)
	check(file)
	if oldKey == file.ObjectKey {
		t.Fatal("retry reused a deleted object")
	}
}
