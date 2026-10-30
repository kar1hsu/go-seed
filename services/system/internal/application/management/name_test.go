package management

import (
	"testing"
	"time"
)

func TestNamesAndBoundedAuditQueries(t *testing.T) {
	for _, v := range []string{"bad\nname", "bad\u202ename", "bad\u200bname", "\xff"} {
		if validName(v) {
			t.Fatalf("invisible/control name accepted: %q", v)
		}
	}
	if !validName("正常 Name") {
		t.Fatal("valid name rejected")
	}
	f, err := auditFilter(AuditQuery{Page: 1, PageSize: 20})
	if err != nil || f.From == nil || f.To == nil || f.To.Sub(*f.From) != 90*24*time.Hour {
		t.Fatal("default audit window unbounded")
	}
	for _, q := range []AuditQuery{{Page: 501, PageSize: 20}, {Page: 1, PageSize: 20, From: "2020-01-01T00:00:00Z", To: "2026-01-01T00:00:00Z"}} {
		if _, err := auditFilter(q); err == nil {
			t.Fatal("unbounded audit query accepted")
		}
	}
}
