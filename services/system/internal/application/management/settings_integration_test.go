//go:build integration

package management

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/resources"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func dictionaryInput() DictionaryInput {
	return DictionaryInput{Name: "订单状态", Enabled: true, Items: []model.DictionaryItem{{Value: "pending", Label: "待处理", Sort: 20, Enabled: true}, {Value: "done", Label: "已完成", Sort: 10, Enabled: true}}}
}
func TestSettingsCRUDPermissionsAndAudit(t *testing.T) {
	f := setup(t)
	readerID, reader := f.user(t, "settings_reader")
	role := f.role(t, "dictionary_reader", resources.DictionaryRead)
	f.assign(t, readerID, role)
	query := Query{Page: 1, PageSize: 20, Enabled: "all"}
	_, err := f.b.CreateDictionary(ctx, reader, "order.status", dictionaryInput())
	code(t, err, codes.PermissionDenied)
	_, _, err = f.b.ListSettings(ctx, reader, query)
	code(t, err, codes.PermissionDenied)
	version, err := repository.NewRBAC(f.db).Version(ctx)
	require(t, err)
	dict, err := f.a.CreateDictionary(ctx, f.root, "order.status", dictionaryInput())
	require(t, err)
	_, err = f.b.CreateDictionary(ctx, f.root, "order.status", dictionaryInput())
	code(t, err, codes.AlreadyExists)
	read, err := f.b.GetDictionary(ctx, reader, dict.Record.ID)
	require(t, err)
	if len(read.Items) != 2 || read.Items[0].Value != "done" {
		t.Fatal("items not persisted/sorted")
	}
	rows, total, err := f.b.ListDictionaries(ctx, reader, query)
	require(t, err)
	if len(rows) != 1 || total != 1 || rows[0].ItemCount != 2 {
		t.Fatal("list summary mismatch")
	}
	query.Keyword = "%"
	_, total, err = f.a.ListDictionaries(ctx, f.root, query)
	require(t, err)
	if total != 0 {
		t.Fatal("search wildcard was not escaped")
	}
	query.Keyword = ""
	code(t, f.a.DeleteDictionary(ctx, f.root, dict.Record.ID, Revision(dict.Record.UpdatedAt)), codes.FailedPrecondition)
	input := dictionaryInput()
	input.Enabled = false
	input.Items = input.Items[:1]
	input.Items[0].Label = "等待处理"
	code(t, f.b.UpdateDictionary(ctx, reader, dict.Record.ID, input, Revision(dict.Record.UpdatedAt)), codes.PermissionDenied)
	require(t, f.a.UpdateDictionary(ctx, f.root, dict.Record.ID, input, Revision(dict.Record.UpdatedAt)))
	code(t, f.b.UpdateDictionary(ctx, f.root, dict.Record.ID, input, Revision(dict.Record.UpdatedAt)), codes.Aborted)
	dict, err = f.a.GetDictionary(ctx, f.root, dict.Record.ID)
	require(t, err)
	if len(dict.Items) != 1 || dict.Items[0].Label != "等待处理" {
		t.Fatal("item replacement lost")
	}
	require(t, f.b.DeleteDictionary(ctx, f.root, dict.Record.ID, Revision(dict.Record.UpdatedAt)))
	_, err = f.a.GetDictionary(ctx, f.root, dict.Record.ID)
	code(t, err, codes.NotFound)
	var remaining int64
	require(t, f.db.Model(&model.DictionaryItem{}).Where("dictionary_id = ?", dict.Record.ID).Count(&remaining).Error)
	if remaining != 0 {
		t.Fatal("orphan dictionary items")
	}
	setting, err := f.a.CreateSetting(ctx, f.root, "page.limit", "integer", SettingInput{Name: "每页数量", Value: "20", Enabled: true})
	require(t, err)
	_, err = f.b.CreateSetting(ctx, f.root, "page.limit", "integer", SettingInput{Name: "重复", Value: "20"})
	code(t, err, codes.AlreadyExists)
	code(t, f.b.UpdateSetting(ctx, f.root, setting.ID, SettingInput{Name: "数量", Value: "wrong"}, Revision(setting.UpdatedAt)), codes.InvalidArgument)
	code(t, f.b.DeleteSetting(ctx, f.root, setting.ID, Revision(setting.UpdatedAt)), codes.FailedPrecondition)
	require(t, f.b.UpdateSetting(ctx, f.root, setting.ID, SettingInput{Name: "每页数量", Value: "50", Enabled: false}, Revision(setting.UpdatedAt)))
	list, total, err := f.a.ListSettings(ctx, f.root, query)
	require(t, err)
	if len(list) != 1 || total != 1 || list[0].Value != "50" || list[0].Type != "integer" {
		t.Fatal("setting update mismatch")
	}
	code(t, f.b.DeleteSetting(ctx, f.root, setting.ID, Revision(setting.UpdatedAt)), codes.Aborted)
	require(t, f.b.DeleteSetting(ctx, f.root, setting.ID, Revision(list[0].UpdatedAt)))
	after, err := repository.NewRBAC(f.db).Version(ctx)
	require(t, err)
	if after != version {
		t.Fatal("ordinary settings must not rebuild policies")
	}
	var logs []model.AuditLog
	require(t, f.db.Where("target_type IN ?", []string{"dictionary", "setting"}).Order("id").Find(&logs).Error)
	if len(logs) != 6 {
		t.Fatalf("audit mutations: %d", len(logs))
	}
	if !strings.Contains(logs[1].ChangesJSON, "等待处理") || !strings.Contains(logs[4].ChangesJSON, "50") {
		t.Fatal("audit snapshots missing")
	}
	for _, log := range logs {
		if log.ActorID != f.rootID {
			t.Fatal("wrong audit actor")
		}
	}
	f.assign(t, readerID)
	_, _, err = f.b.ListDictionaries(ctx, reader, query)
	code(t, err, codes.PermissionDenied)
}
func TestSettingsConcurrentRevisionAndAuditRollback(t *testing.T) {
	f := setup(t)
	dict, err := f.a.CreateDictionary(ctx, f.root, "order.status", dictionaryInput())
	require(t, err)
	in := dictionaryInput()
	in.Enabled = false
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, service := range []*Service{f.a, f.b} {
		wg.Add(1)
		go func(s *Service) {
			defer wg.Done()
			results <- s.UpdateDictionary(ctx, f.root, dict.Record.ID, in, Revision(dict.Record.UpdatedAt))
		}(service)
	}
	wg.Wait()
	close(results)
	counts := map[codes.Code]int{}
	for err := range results {
		counts[status.Code(err)]++
	}
	if counts[codes.OK] != 1 || counts[codes.Aborted] != 1 {
		t.Fatalf("concurrent writes: %v", counts)
	}
	current, err := f.a.GetDictionary(ctx, f.root, dict.Record.ID)
	require(t, err)
	setting, err := f.a.CreateSetting(ctx, f.root, "site.title", "text", SettingInput{Name: "站点", Value: "GoSeed", Enabled: false})
	require(t, err)
	// Force the last write to fail: parents, children, values and deletions must all roll back.
	callback := "test:settings_audit_failure"
	require(t, f.db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "audit_logs" {
			tx.AddError(errors.New("injected audit failure"))
		}
	}))
	t.Cleanup(func() { f.db.Callback().Create().Remove(callback) })
	broken := dictionaryInput()
	broken.Items = broken.Items[:1]
	broken.Name = "should rollback"
	code(t, f.b.UpdateDictionary(ctx, f.root, current.Record.ID, broken, Revision(current.Record.UpdatedAt)), codes.Unavailable)
	code(t, f.b.DeleteDictionary(ctx, f.root, current.Record.ID, Revision(current.Record.UpdatedAt)), codes.Unavailable)
	_, err = f.b.CreateDictionary(ctx, f.root, "failed.create", dictionaryInput())
	code(t, err, codes.Unavailable)
	code(t, f.b.UpdateSetting(ctx, f.root, setting.ID, SettingInput{Name: "changed", Value: "new"}, Revision(setting.UpdatedAt)), codes.Unavailable)
	code(t, f.b.DeleteSetting(ctx, f.root, setting.ID, Revision(setting.UpdatedAt)), codes.Unavailable)
	_, err = f.b.CreateSetting(ctx, f.root, "failed.setting", "boolean", SettingInput{Name: "failed", Value: "true"})
	code(t, err, codes.Unavailable)
	unchanged, err := f.a.GetDictionary(ctx, f.root, current.Record.ID)
	require(t, err)
	if unchanged.Record.Name != current.Record.Name || len(unchanged.Items) != 2 || Revision(unchanged.Record.UpdatedAt) != Revision(current.Record.UpdatedAt) {
		t.Fatal("dictionary audit failure did not roll back")
	}
	list, total, err := f.a.ListSettings(ctx, f.root, Query{Page: 1, PageSize: 20, Enabled: "all"})
	require(t, err)
	if total != 1 || list[0].Value != "GoSeed" || list[0].Name != "站点" {
		t.Fatal("setting audit failure did not roll back")
	}
	_, total, err = f.a.ListDictionaries(ctx, f.root, Query{Page: 1, PageSize: 20, Enabled: "all"})
	require(t, err)
	if total != 1 {
		t.Fatal("failed create persisted")
	}
}
