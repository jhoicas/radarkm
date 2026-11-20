package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Dir resolves the migrations directory (services/api/migrations).
func Dir() (string, error) {
	candidates := []string{
		os.Getenv("RADARKM_MIGRATIONS_DIR"),
		"migrations",
		filepath.Join("services", "api", "migrations"),
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "migrations"),
			filepath.Join(wd, "services", "api", "migrations"),
		)
		dir := wd
		for i := 0; i < 6; i++ {
			candidates = append(candidates, filepath.Join(dir, "migrations"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		info, err := os.Stat(c)
		if err == nil && info.IsDir() {
			abs, err := filepath.Abs(c)
			if err == nil {
				return abs, nil
			}
			return c, nil
		}
	}
	return "", fmt.Errorf("migrations directory not found")
}

func versionFromUpFile(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, ".up.sql")
}

func Up(ctx context.Context, pool *pgxpool.Pool) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, f := range files {
		version := versionFromUpFile(f)
		applied, err := IsApplied(ctx, pool, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("migrate up %s: %w", filepath.Base(f), err)
		}
	}
	return nil
}

func Down(ctx context.Context, pool *pgxpool.Pool) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.down.sql"))
	if err != nil {
		return err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	for _, f := range files {
		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("migrate down %s: %w", filepath.Base(f), err)
		}
	}
	return nil
}

// Reset drops schema objects then applies up migrations.
func Reset(ctx context.Context, pool *pgxpool.Pool) error {
	_ = Down(ctx, pool)
	return Up(ctx, pool)
}

func IsApplied(ctx context.Context, pool *pgxpool.Pool, version string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'schema_migrations'
		)`).Scan(&exists)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	err = pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version,
	).Scan(&exists)
	return exists, err
}

func ListPublicTables(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	rows, err := pool.Query(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

func HasTable(tables []string, name string) bool {
	for _, t := range tables {
		if strings.EqualFold(t, name) {
			return true
		}
	}
	return false
}
