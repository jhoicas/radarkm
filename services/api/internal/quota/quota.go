package quota

import (
	"encoding/json"
	"net/http"
	"time"

	_ "time/tzdata"

	"radarkm/api/internal/rediskeys"
)

// Service is the authoritative free-quota gateway (Redis America/Bogota).
// Consume/business logic belongs to later tasks; T-01 only wires day/key contracts.
type Service struct{}

func NewService() *Service { return &Service{} }

type statusResponse struct {
	DeviceID      string `json:"deviceId"`
	DayBogota     string `json:"dayBogota"`
	QuotaKey      string `json:"quotaKey"`
	Used          int64  `json:"used"`
	Limit         int64  `json:"limit"`
	Remaining     int64  `json:"remaining"`
	Signature     string `json:"signatureEd25519Base64"`
	Authoritative bool   `json:"authoritative"`
}

func (s *Service) HandleStatus(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("deviceId")
	now := time.Now().UTC()
	day := rediskeys.BogotaDay(now)
	resp := statusResponse{
		DeviceID:      deviceID,
		DayBogota:     day,
		QuotaKey:      rediskeys.QuotaKey(deviceID, now),
		Used:          0,
		Limit:         2,
		Remaining:     2,
		Signature:     "",
		Authoritative: true,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Service) HandleConsume(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{
		"error": "quota consume not implemented in this task",
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
