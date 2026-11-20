package rediskeys

import (
	"fmt"
	"time"

	// Embed IANA TZ database so America/Bogota works on Windows and minimal Docker images.
	_ "time/tzdata"
)

const (
	ConfigRuntimeKey = "config:runtime"
	BogotaLocation   = "America/Bogota"
	DedupTTLSeconds  = 90
)

// BogotaDay returns the civil date YYYY-MM-DD in America/Bogota for the given instant.
func BogotaDay(now time.Time) string {
	return now.In(MustBogota()).Format("2006-01-02")
}

// MustBogota loads America/Bogota (tzdata-backed).
func MustBogota() *time.Location {
	loc, err := time.LoadLocation(BogotaLocation)
	if err != nil {
		return time.FixedZone("America/Bogota", -5*3600)
	}
	return loc
}

// NextBogotaMidnight returns the next 00:00:00 America/Bogota strictly after now.
// If now is exactly midnight Bogotá, returns the following day's midnight (for EXPIREAT).
func NextBogotaMidnight(now time.Time) time.Time {
	loc := MustBogota()
	local := now.In(loc)
	midnightToday := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	if local.Before(midnightToday) {
		return midnightToday
	}
	return midnightToday.Add(24 * time.Hour)
}

// QuotaKey builds quota:{device_hash}:{YYYY-MM-DD} for the Bogotá civil day of now.
func QuotaKey(deviceHash string, now time.Time) string {
	return fmt.Sprintf("quota:%s:%s", deviceHash, BogotaDay(now))
}

// QuotaKeyForDay builds quota:{device_hash}:{day} when the Bogotá day is already known.
func QuotaKeyForDay(deviceHash, dayBogota string) string {
	return fmt.Sprintf("quota:%s:%s", deviceHash, dayBogota)
}

// DedupKey builds dedup:{device_hash}:{offer_hash}.
func DedupKey(deviceHash, offerHash string) string {
	return fmt.Sprintf("dedup:%s:%s", deviceHash, offerHash)
}
