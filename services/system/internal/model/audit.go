package model

import "time"

// AuditLog is append-only through the application. Historical names survive role deletion.
type AuditLog struct {
	ID            uint64 `gorm:"primaryKey"`
	Action        string
	ActorID       uint64
	ActorUsername string
	ActorName     string
	Source        string
	TargetType    string
	TargetID      uint64
	TargetKey     string
	TargetName    string
	TraceID       string
	ChangesJSON   string `gorm:"column:changes"`
	CreatedAt     time.Time
}
