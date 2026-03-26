package service

import (
	"context"
	"database/sql"
	"log"
	"time"

	"adnx_dns/internal/errs"
	"adnx_dns/internal/godaddy"
	"adnx_dns/internal/model"
	"adnx_dns/internal/repository"
)

type DomainService struct {
	Repo   *repository.DomainRepository
	Client *godaddy.Client
}

func (s *DomainService) ListAvailable(ctx context.Context) ([]model.Domain, error) {
	return s.Repo.ListByAvailability(ctx, true)
}

func (s *DomainService) ListUnavailable(ctx context.Context) ([]model.Domain, error) {
	return s.Repo.ListByAvailability(ctx, false)
}

func (s *DomainService) ListAvailableDetails(ctx context.Context) ([]model.DomainDetail, error) {
	return s.Repo.ListAvailableDetails(ctx)
}

func (s *DomainService) SetEnabled(ctx context.Context, domain string, enabled bool) error {
	d, err := s.Repo.GetByName(ctx, domain)
	if err != nil {
		if err == sql.ErrNoRows { return errs.New(errs.CodeDomainNotFound, "domain not found") }
		return errs.New(errs.CodeDatabaseError, err.Error())
	}
	if enabled && d.IsAvailable {
		return errs.New(errs.CodeDomainAlreadyOn, "domain already enabled")
	}
	if !enabled && !d.IsAvailable {
		return errs.New(errs.CodeDomainAlreadyOff, "domain already disabled")
	}
	if err := s.Repo.SetAvailability(ctx, domain, enabled); err != nil {
		return errs.New(errs.CodeDatabaseError, err.Error())
	}
	return nil
}

func (s *DomainService) SyncFromGoDaddy(ctx context.Context) (map[string]any, error) {
	items, err := s.Client.ListDomains(ctx)
	if err != nil {
		if _, ok := err.(*godaddy.ErrRateLimited); ok { return nil, errs.New(errs.CodeRateLimited, err.Error()) }
		return nil, errs.New(errs.CodeProviderError, err.Error())
	}
	names := make([]string, 0, len(items))
	for _, it := range items { names = append(names, it.Domain) }
	inserted, updated, err := s.Repo.UpsertFromGoDaddy(ctx, names)
	if err != nil { return nil, errs.New(errs.CodeDatabaseError, err.Error()) }
	return map[string]any{"total_remote": len(items), "inserted": inserted, "updated": updated}, nil
}

func (s *DomainService) StartSyncLoop(ctx context.Context, interval time.Duration) {
	if interval <= 0 { return }
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := s.SyncFromGoDaddy(ctx); err != nil {
					log.Printf("domain sync loop failed: %v", err)
				}
			}
		}
	}()
}
