package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"adnx_dns/internal/errs"
	"adnx_dns/internal/service"
)

type Handler struct {
	Domains  *service.DomainService
	Bindings *service.BindingService
}

type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, payload Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func ok(w http.ResponseWriter, data interface{}, msg string) {
	if msg == "" { msg = "ok" }
	writeJSON(w, http.StatusOK, Response{Code: errs.CodeOK, Data: data, Message: msg})
}

func fail(w http.ResponseWriter, e error) {
	if app, ok := e.(*errs.AppError); ok {
		writeJSON(w, http.StatusOK, Response{Code: app.Code, Data: nil, Message: app.Message})
		return
	}
	writeJSON(w, http.StatusOK, Response{Code: errs.CodeInternal, Data: nil, Message: e.Error()})
}

func decode(r *http.Request, out any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(out)
}

func (h *Handler) GetAvailableDomains(w http.ResponseWriter, r *http.Request) {
	items, err := h.Domains.ListAvailable(r.Context())
	if err != nil { fail(w, err); return }
	ok(w, items, "ok")
}

func (h *Handler) GetAvailableDomainDetails(w http.ResponseWriter, r *http.Request) {
	items, err := h.Domains.ListAvailableDetails(r.Context())
	if err != nil { fail(w, err); return }
	ok(w, items, "ok")
}

func (h *Handler) GetUnavailableDomains(w http.ResponseWriter, r *http.Request) {
	items, err := h.Domains.ListUnavailable(r.Context())
	if err != nil { fail(w, err); return }
	ok(w, items, "ok")
}

func (h *Handler) ResolveIP(w http.ResponseWriter, r *http.Request) {
	var req service.ResolveRequest
	if err := decode(r, &req); err != nil { fail(w, errs.New(errs.CodeInvalidParam, "invalid json")); return }
	resp, err := h.Bindings.Resolve(r.Context(), req)
	if err != nil { fail(w, err); return }
	ok(w, resp, "resolve success")
}

func (h *Handler) GetBindingsByIP(w http.ResponseWriter, r *http.Request) {
	resp, err := h.Bindings.ListByIP(r.Context(), r.URL.Query().Get("ipv4"))
	if err != nil { fail(w, err); return }
	ok(w, resp, "ok")
}

func (h *Handler) Unbind(w http.ResponseWriter, r *http.Request) {
	var req service.UnbindRequest
	if err := decode(r, &req); err != nil { fail(w, errs.New(errs.CodeInvalidParam, "invalid json")); return }
	resp, err := h.Bindings.Unbind(r.Context(), req)
	if err != nil { fail(w, err); return }
	ok(w, resp, "unbind success")
}

func (h *Handler) DisableDomain(w http.ResponseWriter, r *http.Request) {
	var req struct{ Domain string `json:"domain"` }
	if err := decode(r, &req); err != nil { fail(w, errs.New(errs.CodeInvalidParam, "invalid json")); return }
	if strings.TrimSpace(req.Domain) == "" { fail(w, errs.New(errs.CodeInvalidParam, "domain is required")); return }
	if err := h.Domains.SetEnabled(r.Context(), req.Domain, false); err != nil { fail(w, err); return }
	ok(w, map[string]any{"domain": strings.ToLower(strings.TrimSpace(req.Domain)), "enabled": false}, "domain disabled")
}

func (h *Handler) EnableDomain(w http.ResponseWriter, r *http.Request) {
	var req struct{ Domain string `json:"domain"` }
	if err := decode(r, &req); err != nil { fail(w, errs.New(errs.CodeInvalidParam, "invalid json")); return }
	if strings.TrimSpace(req.Domain) == "" { fail(w, errs.New(errs.CodeInvalidParam, "domain is required")); return }
	if err := h.Domains.SetEnabled(r.Context(), req.Domain, true); err != nil { fail(w, err); return }
	ok(w, map[string]any{"domain": strings.ToLower(strings.TrimSpace(req.Domain)), "enabled": true}, "domain enabled")
}

func (h *Handler) SyncDomains(w http.ResponseWriter, r *http.Request) {
	resp, err := h.Domains.SyncFromGoDaddy(r.Context())
	if err != nil { fail(w, err); return }
	ok(w, resp, "sync success")
}
