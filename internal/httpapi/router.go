package httpapi

import "net/http"

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { ok(w, map[string]any{"status": "ok"}, "ok") })
	mux.HandleFunc("/api/v1/domains/available/detail", h.GetAvailableDomainDetails)
	mux.HandleFunc("/api/v1/domains/available", h.GetAvailableDomains)
	mux.HandleFunc("/api/v1/domains/unavailable", h.GetUnavailableDomains)
	mux.HandleFunc("/api/v1/domains/disable", h.DisableDomain)
	mux.HandleFunc("/api/v1/domains/enable", h.EnableDomain)
	mux.HandleFunc("/api/v1/domains/sync", h.SyncDomains)
	mux.HandleFunc("/api/v1/records/resolve", h.ResolveIP)
	mux.HandleFunc("/api/v1/records/by-ip", h.GetBindingsByIP)
	mux.HandleFunc("/api/v1/records/unbind", h.Unbind)
	return mux
}
