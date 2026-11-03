package management

import (
	"context"
	"slices"
	"strconv"
	"strings"

	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/internal/audit"
	"github.com/kar1hsu/go-seed/services/system/internal/authorization"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/resources"
	"google.golang.org/grpc/codes"
)

type menuCatalog struct {
	routes []resources.MenuRoute
	byPath map[string]resources.MenuRoute
	icons  map[string]bool
}

func newMenuCatalog(routes []resources.MenuRoute) menuCatalog {
	c := menuCatalog{routes, make(map[string]resources.MenuRoute), make(map[string]bool)}
	for _, route := range routes {
		c.byPath[route.Path] = route
		c.icons[route.Icon] = true
	}
	return c
}

// Overlay the persisted layout on the current build's registry. New modules appear
// in their declared group even in an installation that already customized its menu.
func (c menuCatalog) reconcile(groups []model.MenuGroup) []model.MenuGroup {
	seen := make(map[string]bool)
	index := make(map[string]int)
	for i := range groups {
		group := &groups[i]
		index[group.ID] = i
		group.Pages = slices.DeleteFunc(group.Pages, func(page model.MenuPage) bool {
			_, registered := c.byPath[page.Path]
			if registered {
				seen[page.Path] = true
			}
			return !registered
		})
	}
	for _, route := range c.routes {
		if seen[route.Path] {
			continue
		}
		id := "builtin-" + route.Group
		i, exists := index[id]
		if !exists {
			i = len(groups)
			index[id] = i
			groups = append(groups, model.MenuGroup{ID: id, Title: route.Group, Icon: route.Icon})
		}
		groups[i].Pages = append(groups[i].Pages, model.MenuPage{Path: route.Path, Title: route.Title, Icon: route.Icon, Hidden: route.Hidden})
	}
	return groups
}

func (c menuCatalog) validate(groups []model.MenuGroup) error {
	invalid := xerr.New(codes.InvalidArgument)
	if len(groups) == 0 || len(groups) > 100 {
		return invalid
	}
	ids, paths := map[string]bool{}, map[string]bool{}
	for _, g := range groups {
		if !validName(g.ID) || strings.TrimSpace(g.ID) != g.ID || ids[g.ID] || !validName(g.Title) || !c.icons[g.Icon] {
			return invalid
		}
		ids[g.ID] = true
		for _, p := range g.Pages {
			if _, ok := c.byPath[p.Path]; !ok || paths[p.Path] || !validName(p.Title) || !c.icons[p.Icon] {
				return invalid
			}
			paths[p.Path] = true
		}
	}
	if len(paths) != len(c.byPath) {
		return invalid
	}
	return nil
}

func (s *Service) loadMenu(ctx context.Context, tx *repository.RBAC) (*model.MenuLayout, error) {
	layout, err := tx.MenuLayout(ctx)
	if err != nil {
		return nil, err
	}
	layout.Groups = s.menus.reconcile(layout.Groups)
	return layout, nil
}

func (s *Service) GetMenuTree(ctx context.Context, raw string) (*model.MenuLayout, error) {
	var layout *model.MenuLayout
	err := s.runData(ctx, raw, resources.MenuRead, false, func(tx *repository.RBAC, _ *authorization.Access) error {
		var err error
		layout, err = s.loadMenu(ctx, tx)
		return err
	})
	return layout, err
}

// Permission checks and navigation use the same session and policy snapshot.
func (s *Service) CurrentNavigation(ctx context.Context, raw string) (*authorization.Access, *model.MenuLayout, error) {
	var actor *authorization.Access
	var layout *model.MenuLayout
	err := s.runData(ctx, raw, "", false, func(tx *repository.RBAC, a *authorization.Access) error {
		var err error
		layout, err = s.loadMenu(ctx, tx)
		if err != nil {
			return err
		}
		actor = a
		for i := range layout.Groups {
			g := &layout.Groups[i]
			g.Pages = slices.DeleteFunc(g.Pages, func(p model.MenuPage) bool {
				perm := s.menus.byPath[p.Path].Permission
				return perm != "" && !a.Has(perm)
			})
		}
		layout.Groups = slices.DeleteFunc(layout.Groups, func(g model.MenuGroup) bool { return len(g.Pages) == 0 })
		return nil
	})
	return actor, layout, err
}

func (s *Service) SaveMenuTree(ctx context.Context, raw, revision string, groups []model.MenuGroup) (*model.MenuLayout, error) {
	var layout *model.MenuLayout
	err := s.runData(ctx, raw, resources.MenuUpdate, true, func(tx *repository.RBAC, a *authorization.Access) error {
		var err error
		layout, err = s.loadMenu(ctx, tx)
		if err != nil {
			return err
		}
		if revision != strconv.FormatUint(layout.Revision, 10) {
			return xerr.New(codes.Aborted)
		}
		if err = s.menus.validate(groups); err != nil {
			return err
		}
		before := audit.Menu(layout.Groups)
		layout.Groups = groups
		layout.Revision++
		if err := tx.SaveMenuLayout(ctx, layout); err != nil {
			return err
		}
		return appendAudit(ctx, tx, a, resources.MenuUpdate, "menu", 1, "navigation", "后台菜单", before, audit.Menu(groups))
	})
	return layout, err
}
