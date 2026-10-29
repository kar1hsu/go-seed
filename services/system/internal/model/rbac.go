package model

import "time"

type Role struct {
	ID          uint64 `gorm:"primaryKey"`
	Code        string
	Name        string
	Description string
	Enabled     bool
	BuiltIn     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
type Permission struct {
	Code      string `gorm:"primaryKey"`
	Name      string
	GroupName string
	Kind      string
	Active    bool
}
type AdminRole struct {
	AdminID uint64 `gorm:"primaryKey"`
	RoleID  uint64 `gorm:"primaryKey"`
}
type RolePermission struct {
	RoleID         uint64 `gorm:"primaryKey"`
	PermissionCode string `gorm:"primaryKey"`
}
type PolicyState struct {
	ID           int `gorm:"primaryKey"`
	Version      uint64
	Bootstrapped bool
}
