package model

import "time"

type Dictionary struct {
	ID                      uint64 `gorm:"primaryKey"`
	Code, Name, Description string
	Enabled                 bool
	CreatedAt, UpdatedAt    time.Time
	ItemCount               int64 `gorm:"->;-:migration"`
}
type DictionaryItem struct {
	DictionaryID uint64 `gorm:"primaryKey;autoIncrement:false"`
	Value        string `gorm:"primaryKey"`
	Label        string
	Sort         int64
	Enabled      bool
}
type Setting struct {
	ID                                   uint64 `gorm:"primaryKey"`
	Code, Name, Description, Type, Value string
	Enabled                              bool
	CreatedAt, UpdatedAt                 time.Time
}
