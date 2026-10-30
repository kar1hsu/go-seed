package config

import (
	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/pkg/storage"
	"github.com/kar1hsu/go-seed/pkg/taskx"
	"github.com/zeromicro/go-zero/core/service"
)

type Config struct {
	AuditRetentionDays int `json:",default=0,range=[0:3650]"`
	service.ServiceConf
	ListenOn    string
	MySQL       dbx.MySQLConfig
	Storage     storage.Config
	Queue       taskx.Config
	Concurrency int `json:",default=4,range=[1:32]"`
}
