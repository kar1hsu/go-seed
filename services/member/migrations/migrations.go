package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var files embed.FS

// Up runs only from the explicit management command, never service startup.
func Up(ctx context.Context, db *sql.DB) error {
	// MySQL DDL auto-commits. A session lock prevents concurrent deployment jobs.
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	var locked int
	if err = conn.QueryRowContext(ctx, "SELECT GET_LOCK('goseed_member_schema', 30)").Scan(&locked); err != nil {
		return err
	}
	if locked != 1 {
		return fmt.Errorf("migration lock unavailable")
	}
	defer conn.ExecContext(context.Background(), "SELECT RELEASE_LOCK('goseed_member_schema')")
	provider, err := goose.NewProvider(goose.DialectMySQL, db, files, goose.WithDisableGlobalRegistry(true))
	if err != nil {
		return err
	}
	_, err = provider.Up(ctx)
	return err
}
