package http

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const readHeaderTimeout = 3 * time.Second

func NewMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.Healthz)
	mux.HandleFunc("GET /readyz", h.Healthz)
	mux.HandleFunc("GET /lookup", h.Lookup)
	mux.HandleFunc("POST /append", h.Append)
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}

func NewServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}
}
