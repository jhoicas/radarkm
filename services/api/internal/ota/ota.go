package ota

import (
	"encoding/json"
	"net/http"
)

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) HandleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"selectors": []any{},
	})
}

func (s *Service) HandleAdminUpsert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": "admin OTA upsert scaffolding",
	})
}
