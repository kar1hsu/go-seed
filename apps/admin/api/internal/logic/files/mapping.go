package files

import (
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
)

func fileRecord(r *pb.FileRecord) types.FileRecord {
	return types.FileRecord{Id: r.Id, Name: r.Name, Size: r.Size, ContentType: r.ContentType, Status: r.Status, Source: r.Source, Sha256: r.Sha256, CreatedAt: r.CreatedAt}
}
func taskRecord(r *pb.TaskRecord) types.TaskRecord {
	return types.TaskRecord{Id: r.Id, Kind: r.Kind, Status: r.Status, Attempts: r.Attempts, MaxAttempts: r.MaxAttempts, Error: r.Error, FileId: r.FileId, TraceId: r.TraceId, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, NextRunAt: r.NextRunAt, Retryable: r.Retryable}
}
