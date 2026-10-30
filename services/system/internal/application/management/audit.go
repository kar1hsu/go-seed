package management

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/internal/audit"
	"github.com/kar1hsu/go-seed/services/system/internal/authorization"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/internal/repository"
	"github.com/kar1hsu/go-seed/services/system/resources"
	"google.golang.org/grpc/codes"
)

func appendAudit(ctx context.Context, tx *repository.RBAC, actor *authorization.Access, action, kind string, id uint64, key, name string, before, after audit.Snapshot) error {
	item, err := audit.New(ctx, actor.Admin, action, kind, id, key, name, before, after)
	if err != nil {
		return err
	}
	return tx.Audit().Append(ctx, item)
}

// Audit stored grants, including retired catalog entries that a replacement or deletion removes.
func assignedPermissions(policy *authorization.Policy, roleID uint64) []string {
	result := []string{}
	for _, grant := range policy.Snapshot.Grants {
		if grant.RoleID == roleID {
			result = append(result, grant.PermissionCode)
		}
	}
	return result
}

type AuditQuery struct {
	Page, PageSize                              int64
	Keyword, Action, ActorID, TraceID, From, To string
}
type AuditDetail struct {
	Record  model.AuditLog
	Changes []audit.Change
}

func auditFilter(q AuditQuery) (repository.AuditFilter, error) {
	f := repository.AuditFilter{Page: q.Page, PageSize: q.PageSize, Keyword: strings.TrimSpace(q.Keyword), Action: q.Action, TraceID: q.TraceID}
	invalid := xerr.New(codes.InvalidArgument)
	if q.Page < 1 || q.PageSize < 1 || q.PageSize > 100 || q.Page > (10000+q.PageSize-1)/q.PageSize || len([]rune(q.Keyword)) > 64 {
		return f, invalid
	}
	if q.Action != "" {
		known := false
		for _, a := range audit.Actions() {
			if a.Code == q.Action {
				known = true
			}
		}
		if !known {
			return f, invalid
		}
	}
	if q.TraceID != "" && !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(q.TraceID) {
		return f, invalid
	}
	if q.ActorID != "" {
		id := uint64(0)
		if q.ActorID != "0" {
			var err error
			id, err = ID(q.ActorID)
			if err != nil {
				return f, err
			}
		}
		f.ActorID = &id
	}
	for _, v := range []struct {
		raw  string
		dest **time.Time
	}{{q.From, &f.From}, {q.To, &f.To}} {
		if v.raw != "" {
			t, err := time.Parse(time.RFC3339Nano, v.raw)
			if err != nil {
				return f, invalid
			}
			t = t.UTC()
			if t.Year() < 1000 || t.Year() > 9999 {
				return f, invalid
			}
			*v.dest = &t
		}
	}
	if f.To == nil {
		v := time.Now().UTC()
		f.To = &v
	}
	if f.From == nil {
		v := f.To.AddDate(0, 0, -90)
		f.From = &v
	}
	if f.To.Sub(*f.From) > 366*24*time.Hour || !f.From.Before(*f.To) {
		return f, invalid
	}
	return f, nil
}
func (s *Service) ListAuditLogs(ctx context.Context, raw string, q AuditQuery) ([]model.AuditLog, int64, error) {
	f, err := auditFilter(q)
	if err != nil {
		return nil, 0, err
	}
	items := []model.AuditLog{}
	var total int64
	err = s.run(ctx, raw, resources.AuditRead, false, func(tx *repository.RBAC, _ *authorization.Access) error {
		var err error
		items, total, err = tx.Audit().List(ctx, f)
		return err
	})
	return items, total, err
}
func (s *Service) GetAuditLog(ctx context.Context, raw string, id uint64) (*AuditDetail, error) {
	if id == 0 {
		return nil, xerr.New(codes.InvalidArgument)
	}
	var result *AuditDetail
	err := s.run(ctx, raw, resources.AuditRead, false, func(tx *repository.RBAC, _ *authorization.Access) error {
		item, err := tx.Audit().Get(ctx, id)
		if err != nil {
			return err
		}
		result = &AuditDetail{Record: *item, Changes: []audit.Change{}}
		return json.Unmarshal([]byte(item.ChangesJSON), &result.Changes)
	})
	return result, err
}
