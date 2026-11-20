package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"radarkm/api/internal/billing"
	"radarkm/api/internal/integrity"
	"radarkm/api/internal/ota"
	"radarkm/api/internal/quota"
)

type Server struct {
	quota     *quota.Service
	integrity *integrity.Service
	billing   *billing.Service
	ota       *ota.Service
}

func New() *Server {
	return &Server{
		quota:     quota.NewService(),
		integrity: integrity.NewService(),
		billing:   billing.NewService(),
		ota:       ota.NewService(),
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", s.handleHealth)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/quota/status", s.quota.HandleStatus)
		r.Post("/quota/consume", s.quota.HandleConsume)
		r.Post("/integrity/verify", s.integrity.HandleVerify)
		r.Get("/ota/selectors", s.ota.HandleList)
		r.Put("/admin/ota/selectors", s.ota.HandleAdminUpsert)
		r.Put("/admin/tariffs", s.billing.HandleAdminTariffs)
		r.Get("/admin/treasury", s.billing.HandleTreasury)
	})

	return r
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "radarkm-api",
	})
}
