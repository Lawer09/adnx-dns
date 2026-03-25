package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"godaddy-dns-sync/internal/config"
	"godaddy-dns-sync/internal/db"
	"godaddy-dns-sync/internal/godaddy"
	"godaddy-dns-sync/internal/httpapi"
	"godaddy-dns-sync/internal/middleware"
	"godaddy-dns-sync/internal/repository"
	"godaddy-dns-sync/internal/service"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		log.Printf("load .env skipped: %v", err)
	}
	sqlDB, err := db.Open(cfg.MySQLDSN)
	if err != nil { log.Fatalf("open mysql failed: %v", err) }
	defer sqlDB.Close()

	gd := godaddy.NewClient(cfg.GoDaddyBaseURL, cfg.GoDaddyAPIKey, cfg.GoDaddyAPISecret, cfg.GoDaddyTimeoutSeconds, cfg.GoDaddyRateLimitPerMinute)
	domainRepo := &repository.DomainRepository{DB: sqlDB}
	bindingRepo := &repository.BindingRepository{DB: sqlDB}
	domainSvc := &service.DomainService{Repo: domainRepo, Client: gd}
	bindSvc := &service.BindingService{Domains: domainRepo, Bindings: bindingRepo, GoDaddy: gd, SubdomainChars: cfg.RandomSubdomainLength}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	domainSvc.StartSyncLoop(ctx, time.Duration(cfg.DomainSyncIntervalSeconds)*time.Second)

	h := &httpapi.Handler{Domains: domainSvc, Bindings: bindSvc}
	router := middleware.RequireAPIToken(cfg.APIToken, httpapi.NewRouter(h))
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: router, ReadHeaderTimeout: 10 * time.Second}

	go func() {
		log.Printf("server listening on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch
	cancel()
	shutdownCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	_ = server.Shutdown(shutdownCtx)
}
