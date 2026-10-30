// Package resources owns the reviewed permission catalog. Runtime policies live in relational tables.
package resources

import (
	_ "embed"
	"fmt"
	"regexp"

	"go.yaml.in/yaml/v3"
)

const (
	FileRead         = "system.file.read"
	FileUpload       = "system.file.upload"
	FileDelete       = "system.file.delete"
	TaskRead         = "system.task.read"
	TaskExport       = "system.task.export"
	TaskRetry        = "system.task.retry"
	DictionaryRead   = "system.dictionary.read"
	DictionaryCreate = "system.dictionary.create"
	DictionaryUpdate = "system.dictionary.update"
	DictionaryDelete = "system.dictionary.delete"
	SettingRead      = "system.setting.read"
	SettingCreate    = "system.setting.create"
	SettingUpdate    = "system.setting.update"
	SettingDelete    = "system.setting.delete"

	UserRead       = "system.user.read"
	UserCreate     = "system.user.create"
	UserUpdate     = "system.user.update"
	UserAssignRole = "system.user.assign_role"
	RoleRead       = "system.role.read"
	RoleCreate     = "system.role.create"
	RoleUpdate     = "system.role.update"
	RoleGrant      = "system.role.grant"
	RoleDelete     = "system.role.delete"
	AuditRead      = "system.audit.read"
	SuperRole      = "super_admin"
)

//go:embed permissions.yaml
var catalog []byte

//go:embed casbin_model.conf
var CasbinModel string

type Permission struct {
	Code     string   `yaml:"code"`
	Name     string   `yaml:"name"`
	Group    string   `yaml:"group"`
	Kind     string   `yaml:"kind"`
	Requires []string `yaml:"requires"`
}

func Catalog() ([]Permission, error) {
	var result []Permission
	if err := yaml.Unmarshal(catalog, &result); err != nil {
		return nil, err
	}
	known := map[string]bool{}
	pattern := regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`)
	for _, p := range result {
		if !pattern.MatchString(p.Code) || known[p.Code] || p.Name == "" || p.Group == "" || (p.Kind != "menu" && p.Kind != "button") {
			return nil, fmt.Errorf("invalid permission catalog entry %q", p.Code)
		}
		known[p.Code] = true
	}
	for _, p := range result {
		for _, dependency := range p.Requires {
			if !known[dependency] || dependency == p.Code {
				return nil, fmt.Errorf("invalid dependency for %s", p.Code)
			}
		}
	}
	return result, nil
}
