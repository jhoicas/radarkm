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
)

// BogotaDay returns the civil date YYYY-MM-DD in America/Bogota for the given instant.
func BogotaDay(now time.Time) string {
	loc, err := time.LoadLocation(BogotaLocation)
	if err != nil {
		// tzdata embed makes this unreachable in normal builds; keep a fixed offset fallback.
		loc = time.FixedZone("America/Bogota", -5*3600)
	}
	return now.In(loc).Format("2006-01-02")
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
