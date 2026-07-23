package common

import (
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var MigrationFiles embed.FS

func RunMigrations() error {
	_, err := DB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := MigrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var versions []int
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		parts := strings.SplitN(e.Name(), "_", 2)
		v, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		versions = append(versions, v)
	}
	sort.Ints(versions)

	for _, v := range versions {
		applied, err := isMigrationApplied(v)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		prefix := fmt.Sprintf("%03d_", v)
		var fileName string
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), prefix) {
				fileName = e.Name()
				break
			}
		}
		if fileName == "" {
			continue
		}

		sql, err := MigrationFiles.ReadFile("migrations/" + fileName)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", fileName, err)
		}

		tx, err := DB.Begin()
		if err != nil {
			return err
		}

		if _, err := tx.Exec(string(sql)); err != nil {
			tx.Rollback()
			return fmt.Errorf("execute migration %s: %w", fileName, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", v); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %d: %w", v, err)
		}

		if err := tx.Commit(); err != nil {
			return err
		}
		fmt.Printf("Migration %s applied\n", fileName)
	}

	return nil
}

func isMigrationApplied(version int) (bool, error) {
	var count int
	err := DB.Get(&count, "SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version)
	return count > 0, err
}
