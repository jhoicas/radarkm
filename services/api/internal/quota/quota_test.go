package quota_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"radarkm/api/internal/db"
	"radarkm/api/internal/migrate"
	"radarkm/api/internal/quota"
	"radarkm/api/internal/rediskeys"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = t
}

func setupQuotaHarness(t *testing.T) (*quota.Service, *db.DeviceRepository, *db.SystemConfigRepository, *db.PlanRepository, *db.EntitlementRepository, *redis.Client, *fakeClock) {
	t.Helper()
	ctx := context.Background()

	pool, err := db.NewPool(ctx, db.DatabaseURL())
	if err != nil {
		t.Fatalf("postgres: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := migrate.Reset(ctx, pool); err != nil {
		t.Fatalf("migrate reset: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr()})
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("redis: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("flush redis: %v", err)
	}

	devices := db.NewDeviceRepository(pool)
	configs := db.NewSystemConfigRepository(pool)
	plans := db.NewPlanRepository(pool)
	ents := db.NewEntitlementRepository(pool)

	clock := &fakeClock{now: bogotaTime(2026, 6, 15, 12, 0, 0)}
	svc := quota.NewService(quota.Deps{
		Redis:         rdb,
		Devices:       devices,
		Configs:       configs,
		Plans:         plans,
		Entitlements:  ents,
		Clock:         clock,
		ConfigCacheTTL: 2 * time.Second,
	})
	return svc, devices, configs, plans, ents, rdb, clock
}

func redisAddr() string {
	return "localhost:6379"
}

func bogotaTime(y int, m time.Month, d, hh, mm, ss int) time.Time {
	return time.Date(y, m, d, hh, mm, ss, 0, rediskeys.MustBogota())
}

func seedDevice(t *testing.T, devices *db.DeviceRepository, hash string, blocked bool) db.Device {
	t.Helper()
	status := "active"
	if blocked {
		status = "blocked"
	}
	dev, err := devices.Insert(context.Background(), db.Device{
		WidevineHash: hash,
		Blocked:      blocked,
		Status:       status,
	})
	if err != nil {
		t.Fatalf("insert device: %v", err)
	}
	return dev
}

func seedPlans(t *testing.T, plans *db.PlanRepository) []db.Plan {
	t.Helper()
	sku := "radarkm.premium.month"
	p1, err := plans.Insert(context.Background(), db.Plan{
		Name: "Pase diario", PlanType: "pass", PriceCop: 4_900, Active: true,
	})
	if err != nil {
		t.Fatalf("plan1: %v", err)
	}
	p2, err := plans.Insert(context.Background(), db.Plan{
		Name: "Premium", PlanType: "subscription", PriceCop: 29_900, PlaySKU: &sku, Active: true,
	})
	if err != nil {
		t.Fatalf("plan2: %v", err)
	}
	_, err = plans.Insert(context.Background(), db.Plan{
		Name: "Archivado", PlanType: "credits", PriceCop: 1_000, Active: false,
	})
	if err != nil {
		t.Fatalf("plan3: %v", err)
	}
	return []db.Plan{p1, p2}
}

func TestConsume_exhaustsLimitThenBlocksWithActivePlans(t *testing.T) {
	svc, devices, configs, plans, _, _, _ := setupQuotaHarness(t)
	ctx := context.Background()
	_, err := configs.UpsertDefault(ctx, 2, "per_device")
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	active := seedPlans(t, plans)
	dev := seedDevice(t, devices, "wv-limit-2", false)

	for i := 0; i < 2; i++ {
		res, err := svc.Consume(ctx, quota.ConsumeRequest{
			DeviceHash:      dev.WidevineHash,
			Platform:        "uber",
			FareCop:         20_000 + int64(i),
			DistanceMeters:  5_000,
		})
		if err != nil {
			t.Fatalf("consume %d: %v", i, err)
		}
		if res.Decision != quota.DecisionAllowed {
			t.Fatalf("consume %d decision=%s", i, res.Decision)
		}
	}

	res, err := svc.Consume(ctx, quota.ConsumeRequest{
		DeviceHash:     dev.WidevineHash,
		Platform:       "uber",
		FareCop:        99_000,
		DistanceMeters: 7_000,
	})
	if err != nil {
		t.Fatalf("third: %v", err)
	}
	if res.Decision != quota.DecisionQuotaExceeded {
		t.Fatalf("want QUOTA_EXCEEDED got %s", res.Decision)
	}
	if res.Remaining != 0 || res.Used != 2 {
		t.Fatalf("used=%d remaining=%d", res.Used, res.Remaining)
	}
	if res.ResetAt.IsZero() {
		t.Fatal("expected ResetAt next Bogotá midnight")
	}
	if len(res.ActivePlans) != len(active) {
		t.Fatalf("active plans=%d want %d", len(res.ActivePlans), len(active))
	}
	for _, p := range res.ActivePlans {
		if !p.Active || p.PriceCop < 0 {
			t.Fatalf("invalid plan %#v", p)
		}
	}
}

func TestConsume_dedup90s_sameOfferDoesNotDiscount(t *testing.T) {
	svc, devices, configs, _, _, _, _ := setupQuotaHarness(t)
	ctx := context.Background()
	_, _ = configs.UpsertDefault(ctx, 2, "per_device")
	dev := seedDevice(t, devices, "wv-dedup", false)

	req := quota.ConsumeRequest{
		DeviceHash:     dev.WidevineHash,
		Platform:       "didi",
		FareCop:        18_000,
		DistanceMeters: 4_200,
	}
	first, err := svc.Consume(ctx, req)
	if err != nil || first.Decision != quota.DecisionAllowed || first.Used != 1 {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	second, err := svc.Consume(ctx, req)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if second.Decision != quota.DecisionDedup {
		t.Fatalf("want DEDUP got %s", second.Decision)
	}
	if second.Used != 1 || second.Remaining != 1 {
		t.Fatalf("dedup must not decrement: used=%d remaining=%d", second.Used, second.Remaining)
	}

	other, err := svc.Consume(ctx, quota.ConsumeRequest{
		DeviceHash:     dev.WidevineHash,
		Platform:       "didi",
		FareCop:        18_500,
		DistanceMeters: 4_200,
	})
	if err != nil || other.Decision != quota.DecisionAllowed || other.Used != 2 {
		t.Fatalf("distinct offer should consume: %+v err=%v", other, err)
	}
}

func TestConsume_dynamicLimitFromSystemConfig_invalidatesRuntimeCache(t *testing.T) {
	svc, devices, configs, _, _, rdb, _ := setupQuotaHarness(t)
	ctx := context.Background()
	cfg, err := configs.UpsertDefault(ctx, 1, "per_device")
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	dev := seedDevice(t, devices, "wv-dyn-limit", false)

	first, err := svc.Consume(ctx, quota.ConsumeRequest{
		DeviceHash: dev.WidevineHash, Platform: "indrive", FareCop: 10_000, DistanceMeters: 1_000,
	})
	if err != nil || first.Decision != quota.DecisionAllowed {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	blocked, err := svc.Consume(ctx, quota.ConsumeRequest{
		DeviceHash: dev.WidevineHash, Platform: "indrive", FareCop: 11_000, DistanceMeters: 1_000,
	})
	if err != nil || blocked.Decision != quota.DecisionQuotaExceeded {
		t.Fatalf("blocked=%+v err=%v", blocked, err)
	}

	if err := configs.UpdateFreeDailyQuota(ctx, cfg.ID, 3); err != nil {
		t.Fatalf("update quota: %v", err)
	}
	if err := svc.InvalidateRuntimeConfig(ctx); err != nil {
		t.Fatalf("invalidate: %v", err)
	}
	if n, _ := rdb.Exists(ctx, rediskeys.ConfigRuntimeKey).Result(); n != 0 {
		t.Fatalf("config:runtime should be gone, exists=%d", n)
	}

	// New Bogotá day key would be needed to raise remaining on same day after already used=1.
	// With limit=3 and used=1, next distinct offer must be ALLOWED.
	ok, err := svc.Consume(ctx, quota.ConsumeRequest{
		DeviceHash: dev.WidevineHash, Platform: "indrive", FareCop: 12_000, DistanceMeters: 1_000,
	})
	if err != nil || ok.Decision != quota.DecisionAllowed || ok.Remaining != 1 {
		t.Fatalf("after limit bump=%+v err=%v", ok, err)
	}
}

func TestConsume_dayRolloverBogota_withInjectableClock(t *testing.T) {
	svc, devices, configs, _, _, _, clock := setupQuotaHarness(t)
	ctx := context.Background()
	_, _ = configs.UpsertDefault(ctx, 2, "per_device")
	dev := seedDevice(t, devices, "wv-day-roll", false)

	clock.Set(bogotaTime(2026, 6, 15, 23, 59, 59))
	for i := 0; i < 2; i++ {
		res, err := svc.Consume(ctx, quota.ConsumeRequest{
			DeviceHash:     dev.WidevineHash,
			Platform:       "cabify",
			FareCop:        15_000 + int64(i),
			DistanceMeters: 3_000,
		})
		if err != nil || res.Decision != quota.DecisionAllowed {
			t.Fatalf("pre-midnight %d: %+v err=%v", i, res, err)
		}
	}
	exceeded, err := svc.Consume(ctx, quota.ConsumeRequest{
		DeviceHash: dev.WidevineHash, Platform: "cabify", FareCop: 16_000, DistanceMeters: 3_000,
	})
	if err != nil || exceeded.Decision != quota.DecisionQuotaExceeded {
		t.Fatalf("still same day: %+v err=%v", exceeded, err)
	}

	clock.Set(bogotaTime(2026, 6, 16, 0, 0, 0))
	fresh, err := svc.Consume(ctx, quota.ConsumeRequest{
		DeviceHash: dev.WidevineHash, Platform: "cabify", FareCop: 17_000, DistanceMeters: 3_000,
	})
	if err != nil || fresh.Decision != quota.DecisionAllowed || fresh.Used != 1 {
		t.Fatalf("new Bogotá day: %+v err=%v", fresh, err)
	}
}

func TestConsume_concurrency_exactlyTwoAllowed(t *testing.T) {
	svc, devices, configs, _, _, _, _ := setupQuotaHarness(t)
	ctx := context.Background()
	_, _ = configs.UpsertDefault(ctx, 2, "per_device")
	dev := seedDevice(t, devices, "wv-race", false)

	const workers = 20
	var allowed, exceeded atomic.Int64
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		i := i
		go func() {
			defer wg.Done()
			res, err := svc.Consume(ctx, quota.ConsumeRequest{
				DeviceHash:     dev.WidevineHash,
				Platform:       "picap",
				FareCop:        10_000 + int64(i),
				DistanceMeters: 2_000 + int64(i),
			})
			if err != nil {
				t.Errorf("worker %d: %v", i, err)
				return
			}
			switch res.Decision {
			case quota.DecisionAllowed:
				allowed.Add(1)
			case quota.DecisionQuotaExceeded:
				exceeded.Add(1)
			default:
				t.Errorf("worker %d unexpected %s", i, res.Decision)
			}
		}()
	}
	wg.Wait()
	if allowed.Load() != 2 || exceeded.Load() != 18 {
		t.Fatalf("allowed=%d exceeded=%d", allowed.Load(), exceeded.Load())
	}
}

func TestConsume_blockedDeviceRejected(t *testing.T) {
	svc, devices, configs, _, _, _, _ := setupQuotaHarness(t)
	ctx := context.Background()
	_, _ = configs.UpsertDefault(ctx, 2, "per_device")
	dev := seedDevice(t, devices, "wv-blocked", true)

	res, err := svc.Consume(ctx, quota.ConsumeRequest{
		DeviceHash: dev.WidevineHash, Platform: "uber", FareCop: 10_000, DistanceMeters: 1_000,
	})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if res.Decision != quota.DecisionDeviceBlocked {
		t.Fatalf("want DEVICE_BLOCKED got %s", res.Decision)
	}
	if res.Used != 0 {
		t.Fatalf("blocked must not consume quota, used=%d", res.Used)
	}
}

func TestConsume_activeEntitlementBypassesQuota(t *testing.T) {
	svc, devices, configs, plans, ents, rdb, clock := setupQuotaHarness(t)
	ctx := context.Background()
	_, _ = configs.UpsertDefault(ctx, 2, "per_device")
	dev := seedDevice(t, devices, "wv-entitled", false)
	plan, err := plans.Insert(ctx, db.Plan{Name: "Premium", PlanType: "subscription", PriceCop: 29_900, Active: true})
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	ends := clock.Now().Add(24 * time.Hour)
	_, err = ents.Insert(ctx, db.Entitlement{
		DeviceID: dev.ID,
		PlanID:   &plan.ID,
		Status:   "active",
		StartsAt: clock.Now().Add(-time.Hour),
		EndsAt:   &ends,
	})
	if err != nil {
		t.Fatalf("entitlement: %v", err)
	}

	for i := 0; i < 5; i++ {
		res, err := svc.Consume(ctx, quota.ConsumeRequest{
			DeviceHash:     dev.WidevineHash,
			Platform:       "uber",
			FareCop:        10_000 + int64(i),
			DistanceMeters: 1_000,
		})
		if err != nil || res.Decision != quota.DecisionAllowed {
			t.Fatalf("entitled consume %d: %+v err=%v", i, res, err)
		}
		if res.EntitlementBypass != true {
			t.Fatalf("expected entitlement bypass flag")
		}
	}
	quotaKey := rediskeys.QuotaKey(dev.WidevineHash, clock.Now())
	if n, _ := rdb.Exists(ctx, quotaKey).Result(); n != 0 {
		t.Fatalf("quota key must not be created under entitlement bypass")
	}
}

func TestOfferHash_isDeterministic(t *testing.T) {
	a := quota.OfferHash("uber", 25_000, 3_500)
	b := quota.OfferHash("uber", 25_000, 3_500)
	c := quota.OfferHash("uber", 25_001, 3_500)
	if a == "" || a != b {
		t.Fatalf("hash not stable: %s vs %s", a, b)
	}
	if a == c {
		t.Fatal("different fare must change hash")
	}
}

func TestBogotaDayContractStillWired(t *testing.T) {
	// Keep HTTP scaffolding compiling against rediskeys day contract.
	_ = fmt.Sprintf("%s", rediskeys.BogotaDay(time.Now()))
}
