package dbx

import (
	"fmt"
	"net"
	"strconv"
	"time"

	driver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MySQLConfig struct {
	Host            string
	Port            int
	Database        string
	User            string
	Password        string
	MaxOpenConns    int           `json:",default=20"`
	MaxIdleConns    int           `json:",default=5"`
	ConnMaxLifetime time.Duration `json:",default=30m"`
}

// OpenMySQL creates a pool without requiring a dependency to be ready at boot.
// Readiness probes report outages; no schema changes run during startup.
func OpenMySQL(c MySQLConfig) (*gorm.DB, error) {
	if c.Host == "" || c.Database == "" || c.User == "" || c.Port < 1 || c.Port > 65535 || c.MaxOpenConns < 1 || c.MaxIdleConns < 0 || c.MaxIdleConns > c.MaxOpenConns || c.ConnMaxLifetime <= 0 {
		return nil, fmt.Errorf("invalid MySQL connection or pool configuration")
	}
	dsn := driver.NewConfig()
	dsn.User, dsn.Passwd, dsn.DBName = c.User, c.Password, c.Database
	dsn.Net, dsn.Addr = "tcp", net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
	dsn.ParseTime, dsn.Loc = true, time.UTC
	dsn.Timeout, dsn.ReadTimeout, dsn.WriteTimeout = 2*time.Second, 2*time.Second, 2*time.Second
	dsn.Params = map[string]string{"charset": "utf8mb4"}
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: dsn.FormatDSN(), SkipInitializeWithVersion: true}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	pool, err := db.DB()
	if err != nil {
		return nil, err
	}
	pool.SetMaxOpenConns(c.MaxOpenConns)
	pool.SetMaxIdleConns(c.MaxIdleConns)
	pool.SetConnMaxLifetime(c.ConnMaxLifetime)
	return db, nil
}
