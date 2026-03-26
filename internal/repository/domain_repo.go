package repository

import (
	"context"
	"database/sql"
	"strings"

	"adnx_dns/internal/model"
)

type DomainRepository struct{ DB *sql.DB }

func scanDomain(scanner interface{ Scan(dest ...any) error }) (*model.Domain, error) {
	var d model.Domain
	var last sql.NullTime
	if err := scanner.Scan(&d.ID, &d.DomainName, &d.Source, &d.SyncStatus, &d.IsAvailable, &last, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return nil, err
	}
	if last.Valid {
		d.LastSyncedAt = &last.Time
	}
	return &d, nil
}

func (r *DomainRepository) ListByAvailability(ctx context.Context, available bool) ([]model.Domain, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, domain_name, source, sync_status, is_available, last_synced_at, created_at, updated_at
		FROM domains WHERE is_available = ? ORDER BY domain_name ASC`, available)
	if err != nil { return nil, err }
	defer rows.Close()
	out := make([]model.Domain, 0)
	for rows.Next() {
		d, err := scanDomain(rows)
		if err != nil { return nil, err }
		out = append(out, *d)
	}
	return out, rows.Err()
}

func (r *DomainRepository) ListAvailableDetails(ctx context.Context) ([]model.DomainDetail, error) {
	domains, err := r.ListByAvailability(ctx, true)
	if err != nil { return nil, err }
	out := make([]model.DomainDetail, 0, len(domains))
	for _, d := range domains {
		recs, err := (&BindingRepository{DB:r.DB}).ListByDomainID(ctx, d.ID, true)
		if err != nil { return nil, err }
		out = append(out, model.DomainDetail{Domain: d, Records: recs})
	}
	return out, nil
}

func (r *DomainRepository) GetByName(ctx context.Context, name string) (*model.Domain, error) {
	row := r.DB.QueryRowContext(ctx, `SELECT id, domain_name, source, sync_status, is_available, last_synced_at, created_at, updated_at
		FROM domains WHERE domain_name = ? LIMIT 1`, strings.ToLower(strings.TrimSpace(name)))
	return scanDomain(row)
}

func (r *DomainRepository) GetAvailableByName(ctx context.Context, name string) (*model.Domain, error) {
	row := r.DB.QueryRowContext(ctx, `SELECT id, domain_name, source, sync_status, is_available, last_synced_at, created_at, updated_at
		FROM domains WHERE domain_name = ? AND is_available = 1 LIMIT 1`, strings.ToLower(strings.TrimSpace(name)))
	return scanDomain(row)
}

func (r *DomainRepository) GetAnyAvailable(ctx context.Context) (*model.Domain, error) {
	row := r.DB.QueryRowContext(ctx, `SELECT id, domain_name, source, sync_status, is_available, last_synced_at, created_at, updated_at
		FROM domains WHERE is_available = 1 ORDER BY id ASC LIMIT 1`)
	return scanDomain(row)
}

func (r *DomainRepository) SetAvailability(ctx context.Context, domain string, enabled bool) error {
	status := "active"
	if !enabled { status = "disabled" }
	res, err := r.DB.ExecContext(ctx, `UPDATE domains SET is_available=?, sync_status=?, updated_at=NOW() WHERE domain_name=?`, enabled, status, strings.ToLower(strings.TrimSpace(domain)))
	if err != nil { return err }
	affected, err := res.RowsAffected()
	if err != nil { return err }
	if affected == 0 { return sql.ErrNoRows }
	return nil
}

func (r *DomainRepository) UpsertFromGoDaddy(ctx context.Context, names []string) (inserted, updated int, err error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil { return 0, 0, err }
	defer tx.Rollback()
	for _, n := range names {
		n = strings.ToLower(strings.TrimSpace(n))
		if n == "" { continue }
		var existing uint64
		_ = tx.QueryRowContext(ctx, `SELECT id FROM domains WHERE domain_name=? LIMIT 1`, n).Scan(&existing)
		_, err = tx.ExecContext(ctx, `INSERT INTO domains(domain_name, source, sync_status, is_available, last_synced_at, created_at, updated_at)
			VALUES (?, 'godaddy', 'active', 1, NOW(), NOW(), NOW())
			ON DUPLICATE KEY UPDATE last_synced_at=NOW(), updated_at=NOW()`, n)
		if err != nil { return 0, 0, err }
		if existing == 0 { inserted++ } else { updated++ }
	}
	if err := tx.Commit(); err != nil { return 0, 0, err }
	return inserted, updated, nil
}
