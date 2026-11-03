package repository

import (
	"context"

	"github.com/kar1hsu/go-seed/services/system/internal/model"
)

func (r *RBAC) MenuLayout(ctx context.Context) (*model.MenuLayout, error) {
	var layout model.MenuLayout
	err := r.db.WithContext(ctx).First(&layout, 1).Error
	return &layout, err
}

// Called inside the same write transaction that locked the policy and checked the revision.
func (r *RBAC) SaveMenuLayout(ctx context.Context, layout *model.MenuLayout) error {
	return r.db.WithContext(ctx).Model(layout).Select("Groups", "Revision").Updates(layout).Error
}
