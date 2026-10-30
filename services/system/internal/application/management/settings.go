package management

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/internal/audit"
	"github.com/kar1hsu/go-seed/services/system/internal/authorization"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/resources"
	"google.golang.org/grpc/codes"
)

type DictionaryInput struct {
	Name, Description string
	Enabled           bool
	Items             []model.DictionaryItem
}
type DictionaryDetail struct {
	Record model.Dictionary
	Items  []model.DictionaryItem
}
type SettingInput struct {
	Name, Description, Value string
	Enabled                  bool
}

func validSettingCode(value string) bool {
	return len(value) <= 64 && regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`).MatchString(value)
}
func validSettingText(value string, max int, empty bool) bool {
	return utf8.ValidString(value) && len([]rune(value)) <= max && (empty || strings.TrimSpace(value) != "") && !strings.ContainsFunc(value, unicode.IsControl)
}
func validateDictionary(in DictionaryInput) error {
	if !validSettingText(in.Name, 64, false) || !validSettingText(in.Description, 256, true) || len(in.Items) > 200 {
		return xerr.New(codes.InvalidArgument)
	}
	seen := map[string]bool{}
	for _, item := range in.Items {
		if !validSettingText(item.Value, 64, false) || item.Value != strings.TrimSpace(item.Value) || !validSettingText(item.Label, 64, false) || item.Sort < 0 || item.Sort > 9999 || seen[item.Value] {
			return xerr.New(codes.InvalidArgument)
		}
		seen[item.Value] = true
	}
	return nil
}
func settingValue(kind string, in SettingInput) (string, error) {
	invalid := xerr.New(codes.InvalidArgument)
	if !validSettingText(in.Name, 64, false) || !validSettingText(in.Description, 256, true) || len(in.Value) > 4096 || !utf8.ValidString(in.Value) {
		return "", invalid
	}
	switch kind {
	case "text":
		return in.Value, nil
	case "integer":
		value, err := strconv.ParseInt(in.Value, 10, 64)
		if err != nil || strconv.FormatInt(value, 10) != in.Value {
			return "", invalid
		}
		return in.Value, nil
	case "boolean":
		if in.Value != "true" && in.Value != "false" {
			return "", invalid
		}
		return in.Value, nil
	case "json":
		var compact bytes.Buffer
		if err := json.Compact(&compact, []byte(in.Value)); err != nil {
			return "", invalid
		}
		return compact.String(), nil
	default:
		return "", invalid
	}
}
func (s *Service) ListDictionaries(ctx context.Context, raw string, q Query) ([]model.Dictionary, int64, error) {
	if err := validateQuery(q); err != nil {
		return nil, 0, err
	}
	items := []model.Dictionary{}
	var total int64
	err := s.runData(ctx, raw, resources.DictionaryRead, false, func(tx *repository.RBAC, _ *authorization.Access) error {
		var err error
		items, total, err = tx.Settings().ListDictionaries(ctx, q.Page, q.PageSize, strings.TrimSpace(q.Keyword), q.Enabled)
		return err
	})
	return items, total, err
}
func (s *Service) GetDictionary(ctx context.Context, raw string, id uint64) (*DictionaryDetail, error) {
	if id == 0 {
		return nil, xerr.New(codes.InvalidArgument)
	}
	var result *DictionaryDetail
	err := s.runData(ctx, raw, resources.DictionaryRead, false, func(tx *repository.RBAC, _ *authorization.Access) error {
		item, err := tx.Settings().Dictionary(ctx, id)
		if err != nil {
			return err
		}
		items, err := tx.Settings().Items(ctx, id)
		if err != nil {
			return err
		}
		item.ItemCount = int64(len(items))
		result = &DictionaryDetail{*item, items}
		return nil
	})
	return result, err
}
func (s *Service) CreateDictionary(ctx context.Context, raw, code string, in DictionaryInput) (*DictionaryDetail, error) {
	if !validSettingCode(code) {
		return nil, xerr.New(codes.InvalidArgument)
	}
	if err := validateDictionary(in); err != nil {
		return nil, err
	}
	now := nextTime(time.Time{})
	item := model.Dictionary{Code: code, Name: strings.TrimSpace(in.Name), Description: strings.TrimSpace(in.Description), Enabled: in.Enabled, CreatedAt: now, UpdatedAt: now, ItemCount: int64(len(in.Items))}
	err := s.runData(ctx, raw, resources.DictionaryCreate, true, func(tx *repository.RBAC, a *authorization.Access) error {
		if err := tx.Settings().CreateDictionary(ctx, &item); err != nil {
			return err
		}
		if err := tx.Settings().ReplaceItems(ctx, item.ID, in.Items); err != nil {
			return err
		}
		return appendAudit(ctx, tx, a, resources.DictionaryCreate, "dictionary", item.ID, item.Code, item.Name, nil, audit.Dictionary(item, in.Items))
	})
	return &DictionaryDetail{item, append([]model.DictionaryItem{}, in.Items...)}, err
}
func (s *Service) UpdateDictionary(ctx context.Context, raw string, id uint64, in DictionaryInput, revision string) error {
	if id == 0 {
		return xerr.New(codes.InvalidArgument)
	}
	if err := validateDictionary(in); err != nil {
		return err
	}
	return s.runData(ctx, raw, resources.DictionaryUpdate, true, func(tx *repository.RBAC, a *authorization.Access) error {
		item, err := tx.Settings().Dictionary(ctx, id)
		if err != nil {
			return err
		}
		if err = revisionMatches(revision, item.UpdatedAt); err != nil {
			return err
		}
		items, err := tx.Settings().Items(ctx, id)
		if err != nil {
			return err
		}
		before := audit.Dictionary(*item, items)
		item.Name, item.Description, item.Enabled, item.UpdatedAt = strings.TrimSpace(in.Name), strings.TrimSpace(in.Description), in.Enabled, nextTime(item.UpdatedAt)
		if err = tx.Settings().UpdateDictionary(ctx, item); err != nil {
			return err
		}
		if err = tx.Settings().ReplaceItems(ctx, id, in.Items); err != nil {
			return err
		}
		return appendAudit(ctx, tx, a, resources.DictionaryUpdate, "dictionary", id, item.Code, item.Name, before, audit.Dictionary(*item, in.Items))
	})
}
func (s *Service) DeleteDictionary(ctx context.Context, raw string, id uint64, revision string) error {
	if id == 0 {
		return xerr.New(codes.InvalidArgument)
	}
	return s.runData(ctx, raw, resources.DictionaryDelete, true, func(tx *repository.RBAC, a *authorization.Access) error {
		item, err := tx.Settings().Dictionary(ctx, id)
		if err != nil {
			return err
		}
		if err = revisionMatches(revision, item.UpdatedAt); err != nil {
			return err
		}
		if item.Enabled {
			return xerr.New(codes.FailedPrecondition)
		}
		items, err := tx.Settings().Items(ctx, id)
		if err != nil {
			return err
		}
		if err = tx.Settings().DeleteDictionary(ctx, id); err != nil {
			return err
		}
		return appendAudit(ctx, tx, a, resources.DictionaryDelete, "dictionary", id, item.Code, item.Name, audit.Dictionary(*item, items), nil)
	})
}
func (s *Service) ListSettings(ctx context.Context, raw string, q Query) ([]model.Setting, int64, error) {
	if err := validateQuery(q); err != nil {
		return nil, 0, err
	}
	items := []model.Setting{}
	var total int64
	err := s.runData(ctx, raw, resources.SettingRead, false, func(tx *repository.RBAC, _ *authorization.Access) error {
		var err error
		items, total, err = tx.Settings().ListSettings(ctx, q.Page, q.PageSize, strings.TrimSpace(q.Keyword), q.Enabled)
		return err
	})
	return items, total, err
}
func (s *Service) CreateSetting(ctx context.Context, raw, code, kind string, in SettingInput) (*model.Setting, error) {
	if !validSettingCode(code) {
		return nil, xerr.New(codes.InvalidArgument)
	}
	value, err := settingValue(kind, in)
	if err != nil {
		return nil, err
	}
	now := nextTime(time.Time{})
	item := model.Setting{Code: code, Name: strings.TrimSpace(in.Name), Description: strings.TrimSpace(in.Description), Type: kind, Value: value, Enabled: in.Enabled, CreatedAt: now, UpdatedAt: now}
	err = s.runData(ctx, raw, resources.SettingCreate, true, func(tx *repository.RBAC, a *authorization.Access) error {
		if err := tx.Settings().CreateSetting(ctx, &item); err != nil {
			return err
		}
		return appendAudit(ctx, tx, a, resources.SettingCreate, "setting", item.ID, item.Code, item.Name, nil, audit.Setting(item))
	})
	return &item, err
}
func (s *Service) UpdateSetting(ctx context.Context, raw string, id uint64, in SettingInput, revision string) error {
	if id == 0 {
		return xerr.New(codes.InvalidArgument)
	}
	return s.runData(ctx, raw, resources.SettingUpdate, true, func(tx *repository.RBAC, a *authorization.Access) error {
		item, err := tx.Settings().Setting(ctx, id)
		if err != nil {
			return err
		}
		if err = revisionMatches(revision, item.UpdatedAt); err != nil {
			return err
		}
		value, err := settingValue(item.Type, in)
		if err != nil {
			return err
		}
		before := audit.Setting(*item)
		item.Name, item.Description, item.Value, item.Enabled, item.UpdatedAt = strings.TrimSpace(in.Name), strings.TrimSpace(in.Description), value, in.Enabled, nextTime(item.UpdatedAt)
		if err = tx.Settings().UpdateSetting(ctx, item); err != nil {
			return err
		}
		return appendAudit(ctx, tx, a, resources.SettingUpdate, "setting", id, item.Code, item.Name, before, audit.Setting(*item))
	})
}
func (s *Service) DeleteSetting(ctx context.Context, raw string, id uint64, revision string) error {
	if id == 0 {
		return xerr.New(codes.InvalidArgument)
	}
	return s.runData(ctx, raw, resources.SettingDelete, true, func(tx *repository.RBAC, a *authorization.Access) error {
		item, err := tx.Settings().Setting(ctx, id)
		if err != nil {
			return err
		}
		if err = revisionMatches(revision, item.UpdatedAt); err != nil {
			return err
		}
		if item.Enabled {
			return xerr.New(codes.FailedPrecondition)
		}
		if err = tx.Settings().DeleteSetting(ctx, id); err != nil {
			return err
		}
		return appendAudit(ctx, tx, a, resources.SettingDelete, "setting", id, item.Code, item.Name, audit.Setting(*item), nil)
	})
}
