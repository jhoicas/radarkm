package integrity

import (
	"encoding/json"
	"net/http"
)

// HandleVerify remains a stub HTTP surface for later tasks (T-03 focuses on Service.Verify).
func (s *Service) HandleVerify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": "HTTP integrity verify wiring deferred; use Service.Verify",
	})
}
