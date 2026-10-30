package model

import "time"

type User struct {
	ID           uint64 `gorm:"primaryKey"`
	Username     string
	Nickname     string
	PasswordHash string
	Enabled      bool
	TokenVersion uint64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
}
type MemberSession struct {
	ID           string `gorm:"primaryKey"`
	UserID       uint64
	TokenVersion uint64
	ExpiresAt    time.Time
	RevokedAt    *time.Time
	CreatedAt    time.Time
}
type MemberRefreshToken struct {
	Digest     string `gorm:"primaryKey"`
	SessionID  string
	ConsumedAt *time.Time
	CreatedAt  time.Time
}
type AuditEvent struct {
	ID        uint64 `gorm:"primaryKey"`
	UserID    uint64
	Action    string
	ActorKind string
	ActorID   uint64
	TraceID   string
	CreatedAt time.Time
}
