package logic

import (
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"
	"github.com/kar1hsu/go-seed/services/system/internal/audit"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"strconv"
)

func auditRecord(item model.AuditLog) *pb.AuditLogRecord {
	return &pb.AuditLogRecord{Id: strconv.FormatUint(item.ID, 10), Action: item.Action, ActionName: audit.ActionName(item.Action), ActorId: strconv.FormatUint(item.ActorID, 10), ActorUsername: item.ActorUsername, ActorName: item.ActorName, Source: item.Source, TargetType: item.TargetType, TargetId: strconv.FormatUint(item.TargetID, 10), TargetKey: item.TargetKey, TargetName: item.TargetName, TraceId: item.TraceID, CreatedAt: management.Revision(item.CreatedAt)}
}
