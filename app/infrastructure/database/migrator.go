package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations applies every "*.up.sql" file in dir, in filename order,
// tracking what has already run in an awebo_schema_migrations table. It's a
// small hand-rolled substitute for golang-migrate so the project doesn't need
// an extra dependency just for this.
//
// The tracking table is deliberately NOT named "schema_migrations" — that's
// the conventional name golang-migrate itself uses (with an incompatible
// bigint version column), and reusing it risks colliding with a database
// that was ever touched by that tool.
func RunMigrations(db *sql.DB, dir string) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS awebo_schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("ensure awebo_schema_migrations table: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".up.sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	for _, name := range files {
		version := strings.TrimSuffix(name, ".up.sql")

		var alreadyApplied int
		if err := db.QueryRow(
			`SELECT COUNT(*) FROM awebo_schema_migrations WHERE version = $1`, version,
		).Scan(&alreadyApplied); err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if alreadyApplied > 0 {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}

		if err := applyMigration(db, version, string(content)); err != nil {
			return err
		}
		fmt.Printf("[migrate] applied %s\n", version)
	}

	return nil
}

func applyMigration(db *sql.DB, version, sqlContent string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx for migration %s: %w", version, err)
	}

	if _, err := tx.Exec(sqlContent); err != nil {
		tx.Rollback()
		return fmt.Errorf("apply migration %s: %w", version, err)
	}

	if _, err := tx.Exec(`INSERT INTO awebo_schema_migrations (version) VALUES ($1)`, version); err != nil {
		tx.Rollback()
		return fmt.Errorf("record migration %s: %w", version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", version, err)
	}

	return nil
}
