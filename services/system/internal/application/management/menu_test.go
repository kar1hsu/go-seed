package management

import (
	"testing"

	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/resources"
)

func TestMenuReconcilesRegisteredPages(t *testing.T) {
	c := newMenuCatalog([]resources.MenuRoute{
		{Path: "/a", Title: "A", Group: "默认", Icon: "Key"},
		{Path: "/new", Title: "新增", Group: "默认", Icon: "Key"},
	})
	groups := []model.MenuGroup{
		{ID: "custom", Title: "自定义", Icon: "Key", Pages: []model.MenuPage{{Path: "/a", Title: "旧设置", Icon: "Key", Hidden: true}, {Path: "/removed", Title: "下线", Icon: "Key"}}},
		{ID: "builtin-默认", Title: "已改名", Icon: "Key"},
	}
	result := c.reconcile(groups)
	if len(result) != 2 || len(result[0].Pages) != 1 || !result[0].Pages[0].Hidden || result[0].Pages[0].Title != "旧设置" || len(result[1].Pages) != 1 || result[1].Pages[0].Path != "/new" || result[1].Title != "已改名" {
		t.Fatalf("unexpected menu after catalog change: %#v", result)
	}
	if err := c.validate(result); err != nil {
		t.Fatal(err)
	}
}

func TestMenuRejectsInvalidTrees(t *testing.T) {
	c := newMenuCatalog([]resources.MenuRoute{{Path: "/only", Title: "Only", Icon: "Menu", Group: "Default"}})
	for name, change := range map[string]func([]model.MenuGroup) []model.MenuGroup{
		"unknown route":   func(g []model.MenuGroup) []model.MenuGroup { g[0].Pages[0].Path = "https://example.com"; return g },
		"missing route":   func(g []model.MenuGroup) []model.MenuGroup { g[0].Pages = nil; return g },
		"duplicate route": func(g []model.MenuGroup) []model.MenuGroup { g[0].Pages = append(g[0].Pages, g[0].Pages[0]); return g },
		"duplicate group ID": func(g []model.MenuGroup) []model.MenuGroup {
			return append(g, model.MenuGroup{ID: g[0].ID, Title: "Copy", Icon: "Menu"})
		},
		"unknown icon":     func(g []model.MenuGroup) []model.MenuGroup { g[0].Icon = "invalid-component"; return g },
		"empty page title": func(g []model.MenuGroup) []model.MenuGroup { g[0].Pages[0].Title = "  "; return g },
	} {
		t.Run(name, func(t *testing.T) {
			if err := c.validate(change(c.reconcile(nil))); err == nil {
				t.Fatal("accepted invalid menu")
			}
		})
	}
}
