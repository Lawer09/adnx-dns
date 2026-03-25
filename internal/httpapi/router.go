package httpapi

import (
	"net/http"
	"strings"
)

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]any{"message":"ok"}) })
	mux.HandleFunc("/api/v1/domains", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListDomains(w, r)
		case http.MethodPost:
			if strings.HasSuffix(r.URL.Path, "/sync") { h.SyncDomains(w, r); return }
			writeJSON(w, 404, map[string]any{"message":"not found"})
		default:
			writeJSON(w, 405, map[string]any{"message":"method not allowed"})
		}
	})
	mux.HandleFunc("/api/v1/domains/sync", h.SyncDomains)
	mux.HandleFunc("/api/v1/bind", h.Bind)
	mux.HandleFunc("/api/v1/unbind", h.Unbind)
	mux.HandleFunc("/api/v1/domains/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/disable") {
			h.DisableDomain(w, r)
			return
		}
		writeJSON(w, 404, map[string]any{"message":"not found"})
	})
	return mux
}
