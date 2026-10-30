package model

import "time"

type File struct {
	ID          uint64 `gorm:"primaryKey"`
	OwnerID     uint64
	Name        string
	ObjectKey   string
	Size        int64
	ContentType string
	Status      string
	Source      string
	SHA256      string `gorm:"column:sha256"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
type Task struct {
	Retryable   bool   `gorm:"-"`
	ID          uint64 `gorm:"primaryKey"`
	OwnerID     uint64
	Kind        string
	DedupKey    string
	Status      string
	Generation  int64
	Attempts    int64
	MaxAttempts int64
	FileID      uint64
	Error       string
	TraceID     string
	LeaseUntil  *time.Time
	NextRunAt   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
type TaskOutbox struct {
	TaskID      uint64 `gorm:"primaryKey"`
	Generation  int64
	AvailableAt time.Time
	DeliveredAt *time.Time
}

func (TaskOutbox) TableName() string { return "task_outbox" }
