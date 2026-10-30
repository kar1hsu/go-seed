// Package audit builds explicit, credential-free business snapshots, never request bodies or ORM dumps.
package audit

import (
	"context"
	"encoding/json"
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"github.com/kar1hsu/go-seed/services/system/resources"
	"github.com/zeromicro/go-zero/core/trace"
)

const PasswordChange = "system.user.change_password"
const CommandDisable = "system.user.command_disable"

type Action struct{ Code, Name string }

func Actions() []Action {
	return []Action{
		{resources.UserCreate, "新增管理员"}, {resources.UserUpdate, "编辑与启停管理员"},
		{resources.UserAssignRole, "分配管理员角色"}, {resources.RoleCreate, "新增角色"},
		{resources.RoleUpdate, "编辑与启停角色"}, {resources.RoleGrant, "设置角色权限"},
		{resources.RoleDelete, "删除角色"}, {PasswordChange, "修改个人密码"}, {CommandDisable, "命令行禁用管理员"},
	}
}
func ActionName(code string) string {
	for _, a := range Actions() {
		if a.Code == code {
			return a.Name
		}
	}
	return code
}

type Field struct {
	Key, Label string
	Values     []string
}
type Snapshot []Field
type Change struct {
	Field  string   `json:"field"`
	Label  string   `json:"label"`
	Before []string `json:"before"`
	After  []string `json:"after"`
}

func Administrator(a model.Administrator) Snapshot {
	return Snapshot{{"username", "用户名", []string{a.Username}}, {"displayName", "显示名称", []string{a.DisplayName}}, {"enabled", "状态", []string{state(a.Enabled)}}}
}
func Role(r model.Role) Snapshot {
	return Snapshot{{"code", "角色标识", []string{r.Code}}, {"name", "角色名称", []string{r.Name}}, {"description", "说明", []string{r.Description}}, {"enabled", "状态", []string{state(r.Enabled)}}}
}
func Roles(roles []model.Role) Snapshot {
	values := []string{}
	for _, role := range roles {
		values = append(values, role.Name+" ("+role.Code+", #"+strconv.FormatUint(role.ID, 10)+")")
	}
	sort.Strings(values)
	return Snapshot{{"roles", "关联角色", values}}
}
func Permissions(permissions []string) Snapshot {
	values := append([]string{}, permissions...)
	sort.Strings(values)
	return Snapshot{{"permissions", "权限", values}}
}
func state(enabled bool) string {
	if enabled {
		return "启用"
	}
	return "停用"
}

func diff(before, after Snapshot) []Change {
	result := []Change{}
	keys := []string{}
	old, next, labels := map[string][]string{}, map[string][]string{}, map[string]string{}
	for _, fields := range []Snapshot{before, after} {
		for _, f := range fields {
			if _, ok := labels[f.Key]; !ok {
				keys = append(keys, f.Key)
			}
			labels[f.Key] = f.Label
		}
	}
	for _, f := range before {
		old[f.Key] = f.Values
	}
	for _, f := range after {
		next[f.Key] = f.Values
	}
	for _, key := range keys {
		if !slices.Equal(old[key], next[key]) {
			result = append(result, Change{key, labels[key], append([]string{}, old[key]...), append([]string{}, next[key]...)})
		}
	}
	return result
}

// New must be persisted on the same transaction handle as the successful business mutation.
// A nil actor denotes the privileged local command, never a caller-provided identity.
func New(ctx context.Context, actor *model.Administrator, action, targetType string, targetID uint64, targetKey, targetName string, before, after Snapshot) (*model.AuditLog, error) {
	changes, err := json.Marshal(diff(before, after))
	if err != nil {
		return nil, err
	}
	item := &model.AuditLog{Action: action, Source: "admin", TargetType: targetType, TargetID: targetID, TargetKey: targetKey, TargetName: targetName, TraceID: trace.TraceIDFromContext(ctx), ChangesJSON: string(changes), CreatedAt: time.Now().UTC()}
	if actor != nil {
		item.ActorID, item.ActorUsername, item.ActorName = actor.ID, actor.Username, actor.DisplayName
	} else {
		item.Source, item.ActorUsername, item.ActorName = "command", "system-cli", "领域运维命令"
	}
	return item, nil
}
