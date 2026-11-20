package db_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"radarkm/api/internal/db"
	"radarkm/api/internal/migrate"
)

func TestSchemaMigrations_UpConstraintsDown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, db.DatabaseURL())
	if err != nil {
		t.Fatalf("postgres unavailable (start docker compose): %v", err)
	}
	defer pool.Close()

	lock, err := migrate.AcquireTestLock(ctx, pool)
	if err != nil {
		t.Fatalf("schema lock: %v", err)
	}
	defer lock.Release(context.Background())

	t.Logf("DATABASE_URL=%s", redactURL(db.DatabaseURL()))

	if err := migrate.ResetLocked(ctx, pool, lock.Conn()); err != nil {
		t.Fatalf("migrate reset/up: %v", err)
	}

	tables, err := migrate.ListPublicTables(ctx, pool)
	if err != nil {
		t.Fatalf("list tables: %v", err)
	}
	required := []string{
		"users", "devices", "system_config", "tariffs", "plans",
		"subscriptions", "entitlements", "transactions",
		"ota_selectors", "integrity_events", "audit_logs", "schema_migrations",
	}
	for _, name := range required {
		if !migrate.HasTable(tables, name) {
			t.Fatalf("missing table %s; have=%v", name, tables)
		}
	}

	applied, err := migrate.IsApplied(ctx, pool, "000001_t01_schema")
	if err != nil || !applied {
		t.Fatalf("schema version not applied: applied=%v err=%v", applied, err)
	}

	assertMoneyColumnsAreBigint(t, ctx, pool)
	assertWidevineUnique(t, ctx, pool)
	assertNegativeAmountBlocked(t, ctx, pool)
	assertAuditLogImmutable(t, ctx, pool)

	if err := migrate.DownLocked(ctx, lock.Conn()); err != nil {
		t.Fatalf("migrate down: %v", err)
	}
	tablesAfter, err := migrate.ListPublicTables(ctx, pool)
	if err != nil {
		t.Fatalf("list tables after down: %v", err)
	}
	for _, name := range required {
		if name == "schema_migrations" {
			continue
		}
		if migrate.HasTable(tablesAfter, name) {
			t.Fatalf("table %s still present after down: %v", name, tablesAfter)
		}
	}

	// Re-apply up so local compose DB stays usable after the suite.
	if err := migrate.UpLocked(ctx, pool, lock.Conn()); err != nil {
		t.Fatalf("migrate up after down: %v", err)
	}
}

func assertMoneyColumnsAreBigint(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	rows, err := pool.Query(ctx, `
		SELECT table_name, column_name, data_type
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND column_name IN ('price_cop', 'amount_cop', 'free_daily_quota', 'device_recall_bits')
		ORDER BY table_name, column_name`)
	if err != nil {
		t.Fatalf("money columns query: %v", err)
	}
	defer rows.Close()
	found := 0
	for rows.Next() {
		var table, col, dtype string
		if err := rows.Scan(&table, &col, &dtype); err != nil {
			t.Fatalf("scan: %v", err)
		}
		found++
		if dtype != "bigint" {
			t.Fatalf("%s.%s type=%s want bigint", table, col, dtype)
		}
	}
	if found < 4 {
		t.Fatalf("expected money/int64 columns, found=%d", found)
	}
}

func assertWidevineUnique(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	repo := db.NewDeviceRepository(pool)
	first, err := repo.Insert(ctx, db.Device{
		WidevineHash:     "wv-unique-test-hash",
		DeviceRecallBits: 1,
		Status:           "active",
	})
	if err != nil {
		t.Fatalf("insert device: %v", err)
	}
	_, err = repo.Insert(ctx, db.Device{
		WidevineHash: "wv-unique-test-hash",
		Status:       "active",
	})
	if err == nil {
		t.Fatal("expected unique violation on widevine_hash")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("want unique_violation 23505, got %v", err)
	}
	got, err := repo.GetByWidevineHash(ctx, "wv-unique-test-hash")
	if err != nil {
		t.Fatalf("get device: %v", err)
	}
	if got.ID != first.ID {
		t.Fatalf("device id mismatch")
	}
}

func assertNegativeAmountBlocked(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	repo := db.NewTransactionRepository(pool)
	_, err := repo.Insert(ctx, db.Transaction{
		Channel:           "play_billing",
		Status:            "pending",
		ExternalReference: "neg-amount-ref",
		AmountCop:         -100,
	})
	if err == nil {
		t.Fatal("expected check violation for negative amount_cop")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23514" {
		t.Fatalf("want check_violation 23514, got %v", err)
	}

	ok, err := repo.Insert(ctx, db.Transaction{
		Channel:           "wompi",
		Status:            "completed",
		ExternalReference: "ok-amount-ref",
		AmountCop:         15_000,
	})
	if err != nil {
		t.Fatalf("insert valid transaction: %v", err)
	}
	if ok.AmountCop != 15_000 {
		t.Fatalf("amount_cop=%d", ok.AmountCop)
	}
}

func assertAuditLogImmutable(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO audit_logs (action, entity_type, payload_after)
		VALUES ('create', 'system_config', '{"free_daily_quota":2}'::jsonb)
		RETURNING id::text`).Scan(&id)
	if err != nil {
		t.Fatalf("insert audit log: %v", err)
	}
	_, err = pool.Exec(ctx, `UPDATE audit_logs SET action = 'tamper' WHERE id = $1::uuid`, id)
	if err == nil {
		t.Fatal("expected audit_logs update to fail")
	}
	_, err = pool.Exec(ctx, `DELETE FROM audit_logs WHERE id = $1::uuid`, id)
	if err == nil {
		t.Fatal("expected audit_logs delete to fail")
	}
}

func redactURL(u string) string {
	if u == "" {
		return u
	}
	return u
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
