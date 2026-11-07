package database

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

const migrationLockID int64 = 723_406_191

// Migrate applies every previously unseen SQL file exactly once. An advisory
// lock serializes concurrent deploys and each file runs in its own transaction.
func Migrate(ctx context.Context, pool *pgxpool.Pool, migrations fs.FS) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, migrationLockID); err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}
	defer conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, migrationLockID)

	var migrationTableExists, legacySchemaExists bool
	if err := conn.QueryRow(ctx, `SELECT to_regclass('public.schema_migrations') IS NOT NULL, to_regclass('public.location') IS NOT NULL`).Scan(&migrationTableExists, &legacySchemaExists); err != nil {
		return fmt.Errorf("inspect migration state: %w", err)
	}
	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		name text PRIMARY KEY,
		applied_at timestamptz NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	entries, err := fs.Glob(migrations, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(entries)
	// Older releases mounted this same directory into docker-entrypoint-initdb.d.
	// Baseline such a database once so an upgrade does not replay destructive seed files.
	if !migrationTableExists && legacySchemaExists {
		for _, name := range entries {
			if _, err := conn.Exec(ctx, `INSERT INTO schema_migrations(name) VALUES ($1)`, filepath.Base(name)); err != nil {
				return fmt.Errorf("baseline legacy migration %s: %w", name, err)
			}
		}
		return nil
	}
	for _, name := range entries {
		var applied bool
		if err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name=$1)`, filepath.Base(name)).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if applied {
			continue
		}

		sql, err := fs.ReadFile(migrations, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if strings.TrimSpace(string(sql)) == "" {
			return fmt.Errorf("migration %s is empty", name)
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}
		if _, err = tx.Exec(ctx, string(sql)); err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO schema_migrations(name) VALUES ($1)`, filepath.Base(name))
		}
		if err == nil {
			err = tx.Commit(ctx)
		} else {
			_ = tx.Rollback(ctx)
		}
		if err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}
	return nil
}
