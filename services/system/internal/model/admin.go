package model

import "time"

type Administrator struct {
	ID           uint64 `gorm:"primaryKey"`
	Username     string
	DisplayName  string
	PasswordHash string
	Enabled      bool
	TokenVersion uint64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
type AdminSession struct {
	ID           string `gorm:"primaryKey"`
	AdminID      uint64
	TokenVersion uint64
	ExpiresAt    time.Time
	RevokedAt    *time.Time
	CreatedAt    time.Time
}
type AdminRefreshToken struct {
	Digest     string `gorm:"primaryKey"`
	SessionID  string
	ConsumedAt *time.Time
	CreatedAt  time.Time
}
