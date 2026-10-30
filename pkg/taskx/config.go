// Package taskx contains queue connection configuration, without domain payloads.
package taskx

import (
	"fmt"
	"github.com/hibiken/asynq"
	"regexp"
)

type Config struct {
	Addr     string
	Password string `json:",optional"`
	DB       int    `json:",default=1"`
	Name     string `json:",default=goseed_dev_system"`
}

func (c Config) Validate() error {
	if c.Addr == "" || c.DB < 0 || !regexp.MustCompile(`^goseed_[a-z0-9_]+_system$`).MatchString(c.Name) {
		return fmt.Errorf("invalid queue configuration")
	}
	return nil
}
func (c Config) Redis() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{Addr: c.Addr, Password: c.Password, DB: c.DB}
}
