package quota

import (
	"testing"
	"time"
)

func TestBogotaDayFormat(t *testing.T) {
	// Fixed instant: 2026-03-20 02:30 UTC → still 2026-03-19 in Bogotá (UTC-5)
	instant := time.Date(2026, 3, 20, 2, 30, 0, 0, time.UTC)
	got := bogotaDay(instant)
	if got != "2026-03-19" {
		t.Fatalf("expected 2026-03-19, got %s", got)
	}
}
