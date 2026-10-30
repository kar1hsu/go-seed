//go:build integration

package management

import (
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/resources"
	"gorm.io/gorm"
	"net/url"
	"strconv"
	"testing"
)

func verifySettingsHTTP(t *testing.T, call func(string, string, string, any, bool, int) map[string]any, db *gorm.DB, prefix, root, reader string) {
	t.Helper()
	dictBody := map[string]any{"code": prefix + "_dict", "name": "HTTP字典", "description": "Smoke fixture", "enabled": true, "items": []map[string]any{{"value": "pending", "label": "待处理", "sort": 10, "enabled": true}}}
	call("GET", "/system/dictionaries", "", nil, false, 401)
	call("GET", "/system/dictionaries", reader, nil, false, 403)
	call("POST", "/system/dictionaries", root, dictBody, false, 403)
	call("POST", "/system/dictionaries", reader, dictBody, true, 403)
	dict := call("POST", "/system/dictionaries", root, dictBody, true, 201)["record"].(map[string]any)
	id := dict["id"].(string)
	path := "/system/dictionaries/" + id
	call("GET", path, reader, nil, false, 403)
	detail := call("GET", path, root, nil, false, 200)
	if len(detail["items"].([]any)) != 1 {
		t.Fatal("HTTP dictionary items lost")
	}
	call("POST", "/system/dictionaries", root, dictBody, true, 409)
	call("DELETE", path+"?revision="+url.QueryEscape(dict["revision"].(string)), root, nil, true, 409)
	update := map[string]any{"name": "停用字典", "description": "Smoke fixture", "enabled": false, "revision": dict["revision"], "items": []any{}}
	call("PUT", path, reader, update, true, 403)
	call("PUT", path, root, update, true, 200)
	call("PUT", path, root, update, true, 409)
	current := call("GET", path, root, nil, false, 200)["record"].(map[string]any)
	list := call("GET", "/system/dictionaries?enabled=disabled&keyword="+prefix, root, nil, false, 200)
	if list["total"].(float64) != 1 || current["itemCount"].(float64) != 0 {
		t.Fatal("dictionary filters/count mismatch")
	}
	call("DELETE", path+"?revision="+url.QueryEscape(current["revision"].(string)), reader, nil, true, 403)
	call("DELETE", path+"?revision="+url.QueryEscape(current["revision"].(string)), root, nil, true, 200)
	call("GET", path, root, nil, false, 404)
	call("GET", "/system/settings", reader, nil, false, 403)
	body := map[string]any{"code": prefix + "_setting", "name": "HTTP配置", "description": "Smoke fixture", "enabled": true, "type": "json", "value": `{"count":9007199254740993}`}
	call("POST", "/system/settings", reader, body, true, 403)
	setting := call("POST", "/system/settings", root, body, true, 201)
	path = "/system/settings/" + setting["id"].(string)
	if setting["value"] != body["value"] {
		t.Fatal("JSON number precision lost")
	}
	update = map[string]any{"name": "HTTP配置", "description": "Smoke fixture", "enabled": false, "revision": setting["revision"], "value": "invalid"}
	call("PUT", path, root, update, true, 400)
	update["value"] = `{"enabled":false}`
	call("PUT", path, reader, update, true, 403)
	call("PUT", path, root, update, false, 403)
	call("PUT", path, root, update, true, 200)
	call("PUT", path, root, update, true, 409)
	list = call("GET", "/system/settings?enabled=disabled&keyword="+prefix, root, nil, false, 200)
	if list["total"].(float64) != 1 {
		t.Fatal("HTTP setting filters mismatch")
	}
	current = list["items"].([]any)[0].(map[string]any)
	call("DELETE", path+"?revision="+url.QueryEscape(current["revision"].(string)), reader, nil, true, 403)
	call("DELETE", path+"?revision="+url.QueryEscape(current["revision"].(string)), root, nil, true, 200)
	var log model.AuditLog
	require(t, db.Where("target_key = ? AND action = ?", prefix+"_setting", resources.SettingUpdate).First(&log).Error)
	history := call("GET", "/system/audit-logs/"+strconv.FormatUint(log.ID, 10), root, nil, false, 200)
	if history["record"].(map[string]any)["targetType"] != "setting" {
		t.Fatal("setting audit missing")
	}
}
