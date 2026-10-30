package repository

import (
	"context"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"gorm.io/gorm"
	"strings"
)

type Settings struct{ db *gorm.DB }

func (r *RBAC) Settings() *Settings { return &Settings{r.db} }
func settingFilter(q *gorm.DB, keyword, enabled string) *gorm.DB {
	if keyword != "" {
		pattern := "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(keyword) + "%"
		q = q.Where("name LIKE ? ESCAPE '!' OR code LIKE ? ESCAPE '!'", pattern, pattern)
	}
	if enabled != "all" {
		q = q.Where("enabled = ?", enabled == "enabled")
	}
	return q
}
func (r *Settings) ListDictionaries(ctx context.Context, page, size int64, keyword, enabled string) ([]model.Dictionary, int64, error) {
	q := settingFilter(r.db.WithContext(ctx).Model(&model.Dictionary{}), keyword, enabled)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := []model.Dictionary{}
	err := q.Select("dictionaries.*, (SELECT COUNT(*) FROM dictionary_items WHERE dictionary_id = dictionaries.id) AS item_count").Order("id DESC").Offset(int((page - 1) * size)).Limit(int(size)).Find(&items).Error
	return items, total, err
}
func (r *Settings) Dictionary(ctx context.Context, id uint64) (*model.Dictionary, error) {
	var item model.Dictionary
	err := r.db.WithContext(ctx).First(&item, id).Error
	return &item, err
}
func (r *Settings) Items(ctx context.Context, id uint64) ([]model.DictionaryItem, error) {
	items := []model.DictionaryItem{}
	err := r.db.WithContext(ctx).Where("dictionary_id = ?", id).Order("sort ASC, value ASC").Find(&items).Error
	return items, err
}
func (r *Settings) CreateDictionary(ctx context.Context, item *model.Dictionary) error {
	return r.db.WithContext(ctx).Create(item).Error
}
func (r *Settings) UpdateDictionary(ctx context.Context, item *model.Dictionary) error {
	return r.db.WithContext(ctx).Model(&model.Dictionary{}).Where("id = ?", item.ID).Updates(map[string]any{"name": item.Name, "description": item.Description, "enabled": item.Enabled, "updated_at": item.UpdatedAt}).Error
}
func (r *Settings) ReplaceItems(ctx context.Context, id uint64, items []model.DictionaryItem) error {
	if err := r.db.WithContext(ctx).Where("dictionary_id = ?", id).Delete(&model.DictionaryItem{}).Error; err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	rows := append([]model.DictionaryItem(nil), items...)
	for i := range rows {
		rows[i].DictionaryID = id
	}
	return r.db.WithContext(ctx).Create(&rows).Error
}
func (r *Settings) DeleteDictionary(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Dictionary{}, id).Error
}
func (r *Settings) ListSettings(ctx context.Context, page, size int64, keyword, enabled string) ([]model.Setting, int64, error) {
	q := settingFilter(r.db.WithContext(ctx).Model(&model.Setting{}), keyword, enabled)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := []model.Setting{}
	err := q.Order("id DESC").Offset(int((page - 1) * size)).Limit(int(size)).Find(&items).Error
	return items, total, err
}
func (r *Settings) Setting(ctx context.Context, id uint64) (*model.Setting, error) {
	var item model.Setting
	err := r.db.WithContext(ctx).First(&item, id).Error
	return &item, err
}
func (r *Settings) CreateSetting(ctx context.Context, item *model.Setting) error {
	return r.db.WithContext(ctx).Create(item).Error
}
func (r *Settings) UpdateSetting(ctx context.Context, item *model.Setting) error {
	return r.db.WithContext(ctx).Model(&model.Setting{}).Where("id = ?", item.ID).Updates(map[string]any{"name": item.Name, "description": item.Description, "value": item.Value, "enabled": item.Enabled, "updated_at": item.UpdatedAt}).Error
}
func (r *Settings) DeleteSetting(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Setting{}, id).Error
}
