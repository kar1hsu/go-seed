// Package menuhttp maps the navigation contract shared by the session and menu endpoints.
package menuhttp

import (
	"context"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/trace"
)

func Tree(in *pb.MenuTree) types.MenuTree {
	tree := types.MenuTree{Revision: in.GetRevision(), Groups: []types.MenuGroup{}}
	for _, g := range in.GetGroups() {
		group := types.MenuGroup{Id: g.Id, Title: g.Title, Icon: g.Icon, Hidden: g.Hidden, Pages: []types.MenuPage{}}
		for _, p := range g.Pages {
			group.Pages = append(group.Pages, types.MenuPage{Path: p.Path, Title: p.Title, Icon: p.Icon, Hidden: p.Hidden})
		}
		tree.Groups = append(tree.Groups, group)
	}
	return tree
}

func Response(ctx context.Context, in *pb.MenuTree, err error) (*types.MenuTreeResponse, error) {
	if err != nil {
		return nil, err
	}
	return &types.MenuTreeResponse{Message: "ok", Data: Tree(in), TraceId: trace.TraceIDFromContext(ctx)}, nil
}

func Request(tree *types.MenuTree) *pb.MenuTree {
	out := &pb.MenuTree{Revision: tree.Revision}
	for _, g := range tree.Groups {
		group := &pb.MenuGroup{Id: g.Id, Title: g.Title, Icon: g.Icon, Hidden: g.Hidden}
		for _, p := range g.Pages {
			group.Pages = append(group.Pages, &pb.MenuPage{Path: p.Path, Title: p.Title, Icon: p.Icon, Hidden: p.Hidden})
		}
		out.Groups = append(out.Groups, group)
	}
	return out
}
