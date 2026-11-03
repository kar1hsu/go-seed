//go:build integration

package management

import (
	"reflect"
	"sync"
	"testing"

	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/resources"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMenuLayoutPersistencePermissionsAndConcurrency(t *testing.T) {
	f := setup(t)
	id, reader := f.user(t, "menu_reader")
	_, err := f.b.GetMenuTree(ctx, reader)
	code(t, err, codes.PermissionDenied)
	f.assign(t, id, f.role(t, "menu_read", resources.MenuRead))
	initial, err := f.a.GetMenuTree(ctx, reader)
	require(t, err)
	_, err = f.b.SaveMenuTree(ctx, reader, "1", initial.Groups)
	code(t, err, codes.PermissionDenied)
	access, nav, err := f.b.CurrentNavigation(ctx, reader)
	require(t, err)
	for _, g := range nav.Groups {
		for _, p := range g.Pages {
			if p.Path == "/system/roles" {
				t.Fatal("unauthorized page in navigation")
			}
		}
	}
	groups := initial.Groups
	dashboard := groups[0].Pages[0]
	groups[0].Pages = groups[0].Pages[1:]
	dashboard.Title = "运营工作台"
	dashboard.Hidden = true
	groups = append([]model.MenuGroup{{ID: "operations", Title: "运营", Icon: "Key", Hidden: true, Pages: []model.MenuPage{dashboard}}}, groups...)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, svc := range []*Service{f.a, f.b} {
		wg.Add(1)
		go func() { defer wg.Done(); _, e := svc.SaveMenuTree(ctx, f.root, "1", groups); results <- e }()
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		switch status.Code(err) {
		case codes.OK:
			success++
		case codes.Aborted:
			conflict++
		default:
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success=%d conflicts=%d", success, conflict)
	}
	saved, err := f.b.GetMenuTree(ctx, reader)
	require(t, err)
	if saved.Revision != 2 || !reflect.DeepEqual(saved.Groups, groups) {
		t.Fatalf("layout not preserved: %#v", saved)
	}
	after, nav, err := f.a.CurrentNavigation(ctx, reader)
	require(t, err)
	if after.Policy.Snapshot.Version != access.Policy.Snapshot.Version || !reflect.DeepEqual(nav.Groups[0], groups[0]) || after.Has(resources.MenuUpdate) {
		t.Fatal("layout changed authorization or failed to reach another session")
	}
	var events int64
	require(t, f.db.Model(&model.AuditLog{}).Where("action = ?", resources.MenuUpdate).Count(&events).Error)
	if events != 1 {
		t.Fatalf("audit events: %d", events)
	}
}
