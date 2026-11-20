package quota

import (
	"encoding/json"
	"net/http"
	"time"

	_ "time/tzdata"

	"radarkm/api/internal/rediskeys"
)

func (s *Service) HandleStatus(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("deviceId")
	now := s.clock.Now()
	day := rediskeys.BogotaDay(now)
	resp := map[string]any{
		"deviceId":      deviceID,
		"dayBogota":     day,
		"quotaKey":      rediskeys.QuotaKey(deviceID, now),
		"used":          int64(0),
		"limit":         int64(2),
		"remaining":     int64(2),
		"signatureEd25519Base64": "",
		"authoritative": true,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Service) HandleConsume(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{
		"error": "HTTP consume wiring deferred; use Service.Consume",
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// keep time import used for documentation of Bogotá day contract.
var _ = time.Time{}
