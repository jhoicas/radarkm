package rediskeys

import (
	"testing"
	"time"
)

func TestBogotaDay_crossesUTCMidnight(t *testing.T) {
	// 2026-03-20 02:30 UTC → still 2026-03-19 in Bogotá (UTC-5)
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
