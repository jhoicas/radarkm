package rediskeys

import (
	"testing"
	"time"
)

func TestBogotaDay_crossesUTCMidnight(t *testing.T) {
	instant := time.Date(2026, 3, 20, 2, 30, 0, 0, time.UTC)
	if got := BogotaDay(instant); got != "2026-03-19" {
		t.Fatalf("BogotaDay=%s want 2026-03-19", got)
	}
}

func TestQuotaKey(t *testing.T) {
	instant := time.Date(2026, 3, 20, 15, 0, 0, 0, time.UTC)
	got := QuotaKey("wv-abc", instant)
	want := "quota:wv-abc:2026-03-20"
	if got != want {
		t.Fatalf("QuotaKey=%s want %s", got, want)
	}
}

func TestDedupAndConfigKeys(t *testing.T) {
	if got := DedupKey("dev1", "offer9"); got != "dedup:dev1:offer9" {
		t.Fatalf("DedupKey=%s", got)
	}
	if ConfigRuntimeKey != "config:runtime" {
		t.Fatalf("ConfigRuntimeKey=%s", ConfigRuntimeKey)
	}
}

func TestNextBogotaMidnight(t *testing.T) {
	now := time.Date(2026, 6, 15, 23, 59, 59, 0, MustBogota())
	next := NextBogotaMidnight(now)
	if next.Year() != 2026 || next.Month() != 6 || next.Day() != 16 || next.Hour() != 0 {
		t.Fatalf("next=%v", next)
	}
}
