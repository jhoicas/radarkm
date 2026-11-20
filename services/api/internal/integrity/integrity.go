package integrity

import (
	"encoding/json"
	"net/http"
)

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) HandleVerify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": "Play Integrity verification scaffolding — implement in later tasks",
	})
}
