//go:build integration

package dbx

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestMigrationPoolOutlivesRequestReadTimeout(t *testing.T) {
	port, err := strconv.Atoi(os.Getenv("MYSQL_PORT"))
	if err != nil {
		t.Skip("requires project MySQL")
	}
	db, err := OpenMigrationMySQL(MySQLConfig{Host: os.Getenv("MYSQL_HOST"), Port: port, Database: os.Getenv("MYSQL_DATABASE"), User: os.Getenv("MYSQL_USER"), Password: os.Getenv("MYSQL_PASSWORD"), MaxOpenConns: 2, MaxIdleConns: 1, ConnMaxLifetime: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.WithContext(ctx).Exec("SELECT SLEEP(2.2)").Error; err != nil {
		t.Fatal("migration still has the API's two-second read timeout", err)
	}
	ctx2, stop := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer stop()
	if err := db.WithContext(ctx2).Exec("SELECT SLEEP(2)").Error; err == nil {
		t.Fatal("migration ignores explicit context deadline")
	}
}

func TestPoolsKeepSQLAndDriverTimeInUTC(t *testing.T) {
	port, err := strconv.Atoi(os.Getenv("MYSQL_PORT"))
	if err != nil {
		t.Skip("requires project MySQL")
	}
	config := MySQLConfig{Host: os.Getenv("MYSQL_HOST"), Port: port, Database: os.Getenv("MYSQL_DATABASE"), User: os.Getenv("MYSQL_USER"), Password: os.Getenv("MYSQL_PASSWORD"), MaxOpenConns: 2, MaxIdleConns: 1, ConnMaxLifetime: time.Minute}
	for _, migration := range []bool{false, true} {
		open := OpenMySQL
		if migration {
			open = OpenMigrationMySQL
		}
		db, err := open(config)
		if err != nil {
			t.Fatal(err)
		}
		pool, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		defer pool.Close()
		var zone string
		var now, utc, roundtrip time.Time
		if err = pool.QueryRow("SELECT @@session.time_zone, NOW(6), UTC_TIMESTAMP(6)").Scan(&zone, &now, &utc); err != nil {
			t.Fatal(err)
		}
		if zone != "+00:00" || !now.Equal(utc) || now.Location() != time.UTC {
			t.Fatalf("SQL clock shifted: zone=%s now=%v utc=%v", zone, now, utc)
		}
		local := time.Date(2026, 11, 1, 1, 30, 0, 0, time.FixedZone("UTC+8", 8*3600))
		if err = pool.QueryRow("SELECT CAST(? AS DATETIME(6))", local).Scan(&roundtrip); err != nil {
			t.Fatal(err)
		}
		if !roundtrip.Equal(local) || roundtrip.Location() != time.UTC {
			t.Fatalf("driver changed the instant: %v => %v", local, roundtrip)
		}
	}
}
