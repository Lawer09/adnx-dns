package service

import (
	"context"
	"database/sql"
	"time"

	"adnx_dns/internal/godaddy"
	"adnx_dns/internal/model"
	"adnx_dns/internal/repository"
)

type DomainService struct {
	Repo   *repository.DomainRepository
	Client *godaddy.Client
}

func (s *DomainService) ListAvailable(ctx context.Context) ([]model.Domain, error) {
	return s.Repo.ListAvailable(ctx)
}

func (s *DomainService) Disable(ctx context.Context, id uint64) error {
	return s.Repo.Disable(ctx, id)
}

func (s *DomainService) SyncFromGoDaddy(ctx context.Context) error {
	domains, err := s.Client.ListDomains(ctx)
	if err != nil { return err }
	names := make([]string, 0, len(domains))
	for _, d := range domains { names = append(names, d.Domain) }
	return s.Repo.UpsertFromGoDaddy(ctx, names)
}

func (s *DomainService) StartSyncLoop(ctx context.Context, interval time.Duration) {
	if interval <= 0 { interval = 5 * time.Minute }
	go func() {
		_ = s.SyncFromGoDaddy(ctx)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = s.SyncFromGoDaddy(context.Background())
			}
		}
	}()
}

func IsNotFound(err error) bool {
	return err == sql.ErrNoRows
}
