package logic

import (
	"github.com/kar1hsu/go-seed/services/system/internal/application/management"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"strconv"
)

func fileRecord(f model.File) *pb.FileRecord {
	return &pb.FileRecord{Id: strconv.FormatUint(f.ID, 10), Name: f.Name, Size: f.Size, ContentType: f.ContentType, Status: f.Status, Source: f.Source, Sha256: f.SHA256, CreatedAt: management.Revision(f.CreatedAt)}
}
func taskRecord(t model.Task) *pb.TaskRecord {
	file := ""
	if t.FileID != 0 {
		file = strconv.FormatUint(t.FileID, 10)
	}
	return &pb.TaskRecord{Id: strconv.FormatUint(t.ID, 10), Kind: t.Kind, Status: t.Status, Attempts: t.Attempts, MaxAttempts: t.MaxAttempts, Error: t.Error, FileId: file, TraceId: t.TraceID, CreatedAt: management.Revision(t.CreatedAt), UpdatedAt: management.Revision(t.UpdatedAt), NextRunAt: management.Revision(t.NextRunAt), Retryable: t.Retryable}
}
