package main

import (
	"log"
	"net/http"
	"os"

	"radarkm/api/internal/httpserver"
)

func main() {
	addr := envOr("RADARKM_API_ADDR", ":8080")
	srv := httpserver.New()
	log.Printf("RadarKm API listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
