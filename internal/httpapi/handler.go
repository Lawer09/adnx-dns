package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"adnx_dns/internal/service"
)

type Handler struct {
	Domains  *service.DomainService
	Bindings *service.BindingService
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) ListDomains(w http.ResponseWriter, r *http.Request) {
	items, err := h.Domains.ListAvailable(r.Context())
	if err != nil { writeJSON(w, 500, map[string]any{"message":err.Error()}); return }
	writeJSON(w, 200, map[string]any{"message":"success", "data":items})
}

func (h *Handler) Bind(w http.ResponseWriter, r *http.Request) {
	var req service.BindRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { writeJSON(w, 400, map[string]any{"message":"invalid json"}); return }
	resp, code, err := h.Bindings.Bind(r.Context(), req)
	if err != nil { writeJSON(w, code, map[string]any{"message":err.Error()}); return }
	writeJSON(w, code, resp)
}

func (h *Handler) DisableDomain(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/domains/")
	idStr = strings.TrimSuffix(idStr, "/disable")
	idStr = strings.Trim(idStr, "/")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil { writeJSON(w, 400, map[string]any{"message":"invalid domain id"}); return }
	if err := h.Domains.Disable(r.Context(), id); err != nil { writeJSON(w, 500, map[string]any{"message":err.Error()}); return }
	writeJSON(w, 200, map[string]any{"message":"domain disabled locally"})
}

func (h *Handler) Unbind(w http.ResponseWriter, r *http.Request) {
	var req service.UnbindRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { writeJSON(w, 400, map[string]any{"message":"invalid json"}); return }
	resp, code, err := h.Bindings.Unbind(r.Context(), req)
	if err != nil { writeJSON(w, code, map[string]any{"message":err.Error()}); return }
	writeJSON(w, code, resp)
}

func (h *Handler) SyncDomains(w http.ResponseWriter, r *http.Request) {
	if err := h.Domains.SyncFromGoDaddy(r.Context()); err != nil {
		code := 500
		if strings.Contains(strings.ToLower(err.Error()), "rate limit") || strings.Contains(err.Error(), "429") { code = 429 }
		writeJSON(w, code, map[string]any{"message":err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"message":"domain sync success"})
}
