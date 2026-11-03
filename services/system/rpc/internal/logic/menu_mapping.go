package logic

import (
	"strconv"

	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
)

func menuTree(layout *model.MenuLayout) *pb.MenuTree {
	out := &pb.MenuTree{Revision: strconv.FormatUint(layout.Revision, 10)}
	for _, g := range layout.Groups {
		group := &pb.MenuGroup{Id: g.ID, Title: g.Title, Icon: g.Icon, Hidden: g.Hidden}
		for _, p := range g.Pages {
			group.Pages = append(group.Pages, &pb.MenuPage{Path: p.Path, Title: p.Title, Icon: p.Icon, Hidden: p.Hidden})
		}
		out.Groups = append(out.Groups, group)
	}
	return out
}

func menuGroups(in []*pb.MenuGroup) []model.MenuGroup {
	groups := make([]model.MenuGroup, 0, len(in))
	for _, g := range in {
		group := model.MenuGroup{ID: g.GetId(), Title: g.GetTitle(), Icon: g.GetIcon(), Hidden: g.GetHidden()}
		for _, p := range g.GetPages() {
			group.Pages = append(group.Pages, model.MenuPage{Path: p.GetPath(), Title: p.GetTitle(), Icon: p.GetIcon(), Hidden: p.GetHidden()})
		}
		groups = append(groups, group)
	}
	return groups
}
