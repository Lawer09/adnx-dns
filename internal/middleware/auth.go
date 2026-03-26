package middleware

import (
	"net/http"
	"strings"

	"adnx_dns/internal/errs"
)

func RequireAPIToken(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		provided := strings.TrimSpace(r.Header.Get("X-API-Token"))
		if provided == "" {
			provided = strings.TrimSpace(r.URL.Query().Get("api_token"))
		}
		if token == "" || provided != token {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"code":1001,"data":null,"message":"invalid api token"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

var _ = errs.CodeInvalidAPIToken
