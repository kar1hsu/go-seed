package config

import (
	"github.com/kar1hsu/go-seed/pkg/taskx"
	"github.com/zeromicro/go-zero/core/service"
)

type Config struct {
	service.ServiceConf
	ListenOn string
	Queue    taskx.Config
}
