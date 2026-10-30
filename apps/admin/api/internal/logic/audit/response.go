package audit

import (
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/services/system/rpc/system"
)

func auditRecord(item *system.AuditLogRecord) types.AuditLogRecord {
	return types.AuditLogRecord{Id: item.Id, Action: item.Action, ActionName: item.ActionName, ActorId: item.ActorId, ActorUsername: item.ActorUsername, ActorName: item.ActorName, Source: item.Source, TargetType: item.TargetType, TargetId: item.TargetId, TargetKey: item.TargetKey, TargetName: item.TargetName, TraceId: item.TraceId, CreatedAt: item.CreatedAt}
}
