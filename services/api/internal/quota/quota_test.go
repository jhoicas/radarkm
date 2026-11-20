package quota

import (
	"testing"
	"time"

	_ "time/tzdata"

	"radarkm/api/internal/rediskeys"
)

func TestBogotaDayViaRedisKeys(t *testing.T) {
	instant := time.Date(2026, 3, 20, 2, 30, 0, 0, time.UTC)
	got := rediskeys.BogotaDay(instant)
	if got != "2026-03-19" {
		t.Fatalf("expected 2026-03-19, got %s", got)
	}
}
