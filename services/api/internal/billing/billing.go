package billing

import (
	"encoding/json"
	"net/http"
)

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) HandleAdminTariffs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": "admin tariffs scaffolding",
	})
}

func (s *Service) HandleTreasury(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"currency": "COP",
		"note":     "All monetary fields use int64/BIGINT",
		"grossCop": int64(0),
		"netCop":   int64(0),
	})
}
