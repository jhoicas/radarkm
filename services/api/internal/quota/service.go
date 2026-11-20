package quota

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"radarkm/api/internal/db"
	"radarkm/api/internal/rediskeys"
)

// Clock is injectable for Bogotá day-rollover tests.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

type Deps struct {
	Redis          redis.Cmdable
	Devices        *db.DeviceRepository
	Configs        *db.SystemConfigRepository
	Plans          *db.PlanRepository
	Entitlements   *db.EntitlementRepository
	Clock          Clock
	ConfigCacheTTL time.Duration
}

type runtimeConfigCache struct {
	FreeDailyQuota    int64  `json:"free_daily_quota"`
	QuotaCountingMode string `json:"quota_counting_mode"`
}

// Service is the authoritative free-quota gateway (Redis America/Bogota + system_config).
type Service struct {
	redis          redis.Cmdable
	devices        *db.DeviceRepository
	configs        *db.SystemConfigRepository
	plans          *db.PlanRepository
	entitlements   *db.EntitlementRepository
	clock          Clock
	configCacheTTL time.Duration
	script         *redis.Script
}

func NewService(deps Deps) *Service {
	clk := deps.Clock
	if clk == nil {
		clk = realClock{}
	}
	ttl := deps.ConfigCacheTTL
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	return &Service{
		redis:          deps.Redis,
		devices:        deps.Devices,
		configs:        deps.Configs,
		plans:          deps.Plans,
		entitlements:   deps.Entitlements,
		clock:          clk,
		configCacheTTL: ttl,
		script:         redis.NewScript(consumeQuotaScript),
	}
}

func (s *Service) Consume(ctx context.Context, req ConsumeRequest) (ConsumeResult, error) {
	if req.DeviceHash == "" {
		return ConsumeResult{}, fmt.Errorf("device hash required")
	}
	if req.FareCop < 0 || req.DistanceMeters < 0 {
		return ConsumeResult{}, fmt.Errorf("fare_cop and distance_meters must be non-negative int64")
	}

	now := s.clock.Now()
	day := rediskeys.BogotaDay(now)
	resetAt := rediskeys.NextBogotaMidnight(now)
	offerHash := OfferHash(req.Platform, req.FareCop, req.DistanceMeters)

	device, err := s.devices.GetByWidevineHash(ctx, req.DeviceHash)
	if err != nil {
		return ConsumeResult{}, fmt.Errorf("device lookup: %w", err)
	}
	if device.Blocked || device.Status == "blocked" || device.Status == "revoked" {
		return ConsumeResult{
			Decision:  DecisionDeviceBlocked,
			DayBogota: day,
			ResetAt:   resetAt,
			OfferHash: offerHash,
		}, nil
	}

	hasEntitlement, err := s.entitlements.HasActive(ctx, device.ID, device.UserID, now)
	if err != nil {
		return ConsumeResult{}, fmt.Errorf("entitlement lookup: %w", err)
	}
	if hasEntitlement {
		cfg, err := s.loadRuntimeConfig(ctx)
		if err != nil {
			return ConsumeResult{}, err
		}
		return ConsumeResult{
			Decision:          DecisionAllowed,
			Used:              0,
			Remaining:         cfg.FreeDailyQuota,
			Limit:             cfg.FreeDailyQuota,
			DayBogota:         day,
			ResetAt:           resetAt,
			EntitlementBypass: true,
			OfferHash:         offerHash,
		}, nil
	}

	cfg, err := s.loadRuntimeConfig(ctx)
	if err != nil {
		return ConsumeResult{}, err
	}

	quotaKey := rediskeys.QuotaKeyForDay(req.DeviceHash, day)
	dedupKey := rediskeys.DedupKey(req.DeviceHash, offerHash)

	// TTL relative to injectable Clock (EXPIREAT would fight Redis wall-clock in tests).
	quotaTTL := int64(resetAt.Sub(now).Seconds())
	if quotaTTL < 1 {
		quotaTTL = 1
	}

	raw, err := s.script.Run(ctx, s.redis, []string{quotaKey, dedupKey},
		cfg.FreeDailyQuota,
		quotaTTL,
		rediskeys.DedupTTLSeconds,
	).Result()
	if err != nil {
		return ConsumeResult{}, fmt.Errorf("quota lua: %w", err)
	}
	decision, used, remaining, err := parseLuaResult(raw)
	if err != nil {
		return ConsumeResult{}, err
	}

	result := ConsumeResult{
		Decision:  decision,
		Used:      used,
		Remaining: remaining,
		Limit:     cfg.FreeDailyQuota,
		DayBogota: day,
		ResetAt:   resetAt,
		OfferHash: offerHash,
	}

	if decision == DecisionQuotaExceeded {
		plans, err := s.plans.ListActive(ctx)
		if err != nil {
			return ConsumeResult{}, fmt.Errorf("list plans: %w", err)
		}
		result.ActivePlans = plans
		result.Remaining = 0
	}
	return result, nil
}

func (s *Service) InvalidateRuntimeConfig(ctx context.Context) error {
	return s.redis.Del(ctx, rediskeys.ConfigRuntimeKey).Err()
}

func (s *Service) loadRuntimeConfig(ctx context.Context) (runtimeConfigCache, error) {
	if cached, err := s.redis.Get(ctx, rediskeys.ConfigRuntimeKey).Bytes(); err == nil {
		var cfg runtimeConfigCache
		if json.Unmarshal(cached, &cfg) == nil && cfg.FreeDailyQuota >= 0 {
			return cfg, nil
		}
	}

	dbCfg, err := s.configs.GetLatest(ctx)
	if err != nil {
		return runtimeConfigCache{}, fmt.Errorf("system_config: %w", err)
	}
	cfg := runtimeConfigCache{
		FreeDailyQuota:    dbCfg.FreeDailyQuota,
		QuotaCountingMode: dbCfg.QuotaCountingMode,
	}
	payload, _ := json.Marshal(cfg)
	_ = s.redis.Set(ctx, rediskeys.ConfigRuntimeKey, payload, s.configCacheTTL).Err()
	return cfg, nil
}

func parseLuaResult(raw any) (Decision, int64, int64, error) {
	arr, ok := raw.([]any)
	if !ok || len(arr) != 3 {
		return "", 0, 0, fmt.Errorf("unexpected lua result: %#v", raw)
	}
	decision, _ := arr[0].(string)
	used, err1 := strconv.ParseInt(fmt.Sprint(arr[1]), 10, 64)
	remaining, err2 := strconv.ParseInt(fmt.Sprint(arr[2]), 10, 64)
	if err1 != nil || err2 != nil {
		return "", 0, 0, fmt.Errorf("parse lua ints: %v %v", err1, err2)
	}
	switch Decision(decision) {
	case DecisionAllowed, DecisionDedup, DecisionQuotaExceeded:
		return Decision(decision), used, remaining, nil
	default:
		return "", 0, 0, fmt.Errorf("unknown lua decision %q", decision)
	}
}

// OfferHash builds a deterministic hash from platform + fare_cop + distance_meters.
func OfferHash(platform string, fareCop, distanceMeters int64) string {
	payload := platform + "|" + strconv.FormatInt(fareCop, 10) + "|" + strconv.FormatInt(distanceMeters, 10)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:16])
}
