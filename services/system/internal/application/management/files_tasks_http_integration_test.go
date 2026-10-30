//go:build integration

package management

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

func verifyFilesTasksHTTP(t *testing.T, call func(string, string, string, any, bool, int) map[string]any, db *gorm.DB, prefix, root, reader string) {
	t.Helper()
	base := os.Getenv("RBAC_SMOKE_URL") + "/admin-api/v1"
	client := &http.Client{Timeout: 40 * time.Second}
	transfer := func(method, path, token string, data []byte, csrf bool, want int) []byte {
		t.Helper()
		req, err := http.NewRequest(method, base+path, bytes.NewReader(data))
		require(t, err)
		req.Header.Set("Content-Type", "application/octet-stream")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if csrf {
			req.Header.Set("X-GoSeed-CSRF", "1")
		}
		resp, err := client.Do(req)
		require(t, err)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		require(t, err)
		if resp.StatusCode != want {
			t.Fatalf("%s %s: status %d want %d", method, path, resp.StatusCode, want)
		}
		if method == "GET" && want == 200 && (resp.Header.Get("Content-Type") != "application/octet-stream" || resp.Header.Get("X-Content-Type-Options") != "nosniff" || resp.Header.Get("Content-Disposition") == "") {
			t.Fatal("private download headers missing")
		}
		return body
	}
	call("GET", "/system/files", "", nil, false, 401)
	call("GET", "/system/tasks", reader, nil, false, 403)
	call("POST", "/system/files", root, map[string]any{"name": "unsafe.svg", "size": 10, "contentType": "image/svg+xml"}, true, 400)
	payload := bytes.Repeat([]byte("GoSeed streaming upload smoke test\n"), 16000)
	body := map[string]any{"name": prefix + ".txt", "size": len(payload), "contentType": "text/plain"}
	call("POST", "/system/files", reader, body, true, 403)
	call("POST", "/system/files", root, body, false, 403)
	record := call("POST", "/system/files", root, body, true, 201)["record"].(map[string]any)
	id := record["id"].(string)
	path := "/system/files/" + id + "/content"
	transfer("PUT", path, root, payload, false, 403)
	transfer("PUT", path, reader, payload, true, 403)
	transfer("PUT", path, root, payload, true, 200)
	transfer("GET", path, "", nil, false, 401)
	transfer("GET", path, reader, nil, false, 403)
	if !bytes.Equal(transfer("GET", path, root, nil, false, 200), payload) {
		t.Fatal("streamed bytes changed")
	}
	listed := call("GET", "/system/files?keyword="+prefix, root, nil, false, 200)
	if listed["total"].(float64) != 1 {
		t.Fatal("file list mismatch")
	}
	item := listed["items"].([]any)[0].(map[string]any)
	if item["sha256"] == "" || item["status"] != "ready" {
		t.Fatal("file not verified")
	}
	if _, ok := item["objectKey"]; ok {
		t.Fatal("private object key exposed")
	}
	var uuid [16]byte
	_, err := rand.Read(uuid[:])
	require(t, err)
	uuid[6] = (uuid[6] & 15) | 64
	uuid[8] = (uuid[8] & 63) | 128
	key := fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
	export := map[string]string{"requestKey": key}
	call("POST", "/system/tasks/dictionary-export", reader, export, true, 403)
	call("POST", "/system/tasks/dictionary-export", root, export, false, 403)
	job := call("POST", "/system/tasks/dictionary-export", root, export, true, 202)["record"].(map[string]any)
	again := call("POST", "/system/tasks/dictionary-export", root, export, true, 202)["record"].(map[string]any)
	if job["id"] != again["id"] {
		t.Fatal("duplicate HTTP export")
	}
	wait := func(id, state string) {
		t.Helper()
		deadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(deadline) {
			var row model.Task
			require(t, db.First(&row, id).Error)
			if row.Status == state {
				return
			}
			if row.Status == "failed" {
				t.Fatalf("task failed: %s", row.Error)
			}
			time.Sleep(150 * time.Millisecond)
		}
		t.Fatal("worker did not reach " + state)
	}
	wait(job["id"].(string), "succeeded")
	data := transfer("GET", "/system/files/"+job["fileId"].(string)+"/content", root, nil, false, 200)
	if !bytes.Contains(data, []byte("字典标识")) {
		t.Fatal("CSV missing header")
	}
	call("POST", "/system/tasks/"+job["id"].(string)+"/retry", root, nil, true, 409)
	for _, fileID := range []string{id, job["fileId"].(string)} {
		call("DELETE", "/system/files/"+fileID, root, nil, true, 202)
		var deletion model.Task
		require(t, db.Where("dedup_key = ?", "file-delete:"+fileID).First(&deletion).Error)
		wait(strconv.FormatUint(deletion.ID, 10), "succeeded")
		transfer("GET", "/system/files/"+fileID+"/content", root, nil, false, 409)
	}
}
