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

const schemaAdvisoryLockKey int64 = 72420101

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

// Lock is a session-scoped PG advisory lock on a dedicated pool connection.
type Lock struct {
	conn *pgxpool.Conn
}

// AcquireTestLock serializes migrate + integration assertions across packages.
func AcquireTestLock(ctx context.Context, pool *pgxpool.Pool) (*Lock, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire conn: %w", err)
	}
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, schemaAdvisoryLockKey); err != nil {
		conn.Release()
		return nil, fmt.Errorf("advisory lock: %w", err)
	}
	return &Lock{conn: conn}, nil
}

// Conn returns the locked connection (for ResetLocked/UpLocked/DownLocked).
func (l *Lock) Conn() *pgxpool.Conn {
	if l == nil {
		return nil
	}
	return l.conn
}

// Release unlocks and returns the connection to the pool.
func (l *Lock) Release(ctx context.Context) {
	if l == nil || l.conn == nil {
		return
	}
	_, _ = l.conn.Exec(ctx, `SELECT pg_advisory_unlock($1)`, schemaAdvisoryLockKey)
	l.conn.Release()
	l.conn = nil
}

func withSchemaLock(ctx context.Context, pool *pgxpool.Pool, fn func(*pgxpool.Conn) error) error {
	lock, err := AcquireTestLock(ctx, pool)
	if err != nil {
		return err
	}
	defer lock.Release(context.Background())
	return fn(lock.conn)
}

// Up applies pending *.up.sql migrations.
func Up(ctx context.Context, pool *pgxpool.Pool) error {
	return withSchemaLock(ctx, pool, func(conn *pgxpool.Conn) error {
		return UpLocked(ctx, pool, conn)
	})
}

// UpLocked applies pending ups using an already-locked connection.
func UpLocked(ctx context.Context, pool *pgxpool.Pool, conn *pgxpool.Conn) error {
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
		applied, err := isAppliedOn(ctx, conn, version)
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
		if _, err := conn.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("migrate up %s: %w", filepath.Base(f), err)
		}
	}
	return nil
}

// Down applies all *.down.sql migrations (newest first).
func Down(ctx context.Context, pool *pgxpool.Pool) error {
	return withSchemaLock(ctx, pool, func(conn *pgxpool.Conn) error {
		return DownLocked(ctx, conn)
	})
}

// DownLocked applies downs using an already-locked connection.
func DownLocked(ctx context.Context, conn *pgxpool.Conn) error {
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
		if _, err := conn.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("migrate down %s: %w", filepath.Base(f), err)
		}
	}
	return nil
}

// Reset drops schema objects then applies up migrations under one session lock.
func Reset(ctx context.Context, pool *pgxpool.Pool) error {
	return withSchemaLock(ctx, pool, func(conn *pgxpool.Conn) error {
		return ResetLocked(ctx, pool, conn)
	})
}

// ResetLocked drops then re-applies ups on an already-locked connection.
func ResetLocked(ctx context.Context, pool *pgxpool.Pool, conn *pgxpool.Conn) error {
	_ = DownLocked(ctx, conn)
	dir, err := Dir()
	if err != nil {
		return err
	}
	upFiles, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return err
	}
	sort.Strings(upFiles)
	for _, f := range upFiles {
		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := conn.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("migrate up %s: %w", filepath.Base(f), err)
		}
	}
	return nil
}

func isAppliedOn(ctx context.Context, conn *pgxpool.Conn, version string) (bool, error) {
	var exists bool
	err := conn.QueryRow(ctx, `
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
	err = conn.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version,
	).Scan(&exists)
	return exists, err
}

// IsApplied reports whether a migration version is recorded.
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

// ListPublicTables returns public BASE TABLE names.
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

// HasTable reports whether name is in tables (case-insensitive).
func HasTable(tables []string, name string) bool {
	for _, t := range tables {
		if strings.EqualFold(t, name) {
			return true
		}
	}
	return false
}
