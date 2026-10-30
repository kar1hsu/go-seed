package management

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"testing"
)

func TestAuditQueryValidation(t *testing.T) {
	valid := AuditQuery{Page: 1, PageSize: 20}
	for name, change := range map[string]func(*AuditQuery){
		"page":    func(q *AuditQuery) { q.Page = 0 },
		"size":    func(q *AuditQuery) { q.PageSize = 101 },
		"keyword": func(q *AuditQuery) { q.Keyword = strings.Repeat("界", 65) },
		"action":  func(q *AuditQuery) { q.Action = "arbitrary" },
		"actor":   func(q *AuditQuery) { q.ActorID = "01" },
		"trace":   func(q *AuditQuery) { q.TraceID = "arbitrary" },
		"date":    func(q *AuditQuery) { q.From = "yesterday" },
		"range":   func(q *AuditQuery) { q.From = "2026-10-30T00:00:00Z"; q.To = q.From },
	} {
		t.Run(name, func(t *testing.T) {
			q := valid
			change(&q)
			_, err := auditFilter(q)
			if status.Code(err) != codes.InvalidArgument {
				t.Fatal(err)
			}
		})
	}
	if _, err := auditFilter(valid); err != nil {
		t.Fatal(err)
	}
}
