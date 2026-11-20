package quota

import (
	"encoding/json"
	"net/http"
	"time"
)

// Service is the authoritative free-quota gateway (Redis America/Bogota in later tasks).
type Service struct{}

func NewService() *Service { return &Service{} }

type statusResponse struct {
	DeviceID   string `json:"deviceId"`
	DayBogota  string `json:"dayBogota"`
	Used       int64  `json:"used"`
	Limit      int64  `json:"limit"`
	Remaining  int64  `json:"remaining"`
	Signature  string `json:"signatureEd25519Base64"`
	Authoritative bool `json:"authoritative"`
}

func (s *Service) HandleStatus(w http.ResponseWriter, r *http.Request) {
	day := bogotaDay(time.Now())
	resp := statusResponse{
		DeviceID:      r.URL.Query().Get("deviceId"),
		DayBogota:     day,
		Used:          0,
		Limit:         2,
		Remaining:     2,
		Signature:     "", // Ed25519 signing wired later
		Authoritative: true,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Service) HandleConsume(w http.ResponseWriter, r *http.Request) {
	// Scaffolding: Integrity + Redis INCR under America/Bogota day key.
	writeJSON(w, http.StatusNotImplemented, map[string]string{
		"error": "quota consume not implemented in scaffolding",
	})
}

func bogotaDay(now time.Time) string {
	loc, err := time.LoadLocation("America/Bogota")
	if err != nil {
		loc = time.FixedZone("America/Bogota", -5*3600)
	}
	return now.In(loc).Format("2006-01-02")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
