package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"godaddy-dns-sync/internal/model"
)

type DomainRepository struct{ DB *sql.DB }

func (r *DomainRepository) ListAvailable(ctx context.Context) ([]model.Domain, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id, domain_name, source, sync_status, is_available, last_synced_at, created_at, updated_at
		FROM domains WHERE is_available = 1 AND sync_status = 'active' ORDER BY domain_name ASC`)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []model.Domain
	for rows.Next() {
		var d model.Domain
		var last sql.NullTime
		if err := rows.Scan(&d.ID,&d.DomainName,&d.Source,&d.SyncStatus,&d.IsAvailable,&last,&d.CreatedAt,&d.UpdatedAt); err != nil { return nil, err }
		if last.Valid { d.LastSyncedAt = &last.Time }
		out = append(out,d)
	}
	return out, rows.Err()
}

func (r *DomainRepository) GetAvailableByName(ctx context.Context, name string) (*model.Domain, error) {
	var d model.Domain
	var last sql.NullTime
	err := r.DB.QueryRowContext(ctx, `SELECT id, domain_name, source, sync_status, is_available, last_synced_at, created_at, updated_at
		FROM domains WHERE domain_name = ? AND is_available = 1 AND sync_status = 'active' LIMIT 1`, strings.ToLower(strings.TrimSpace(name))).
		Scan(&d.ID,&d.DomainName,&d.Source,&d.SyncStatus,&d.IsAvailable,&last,&d.CreatedAt,&d.UpdatedAt)
	if err != nil { return nil, err }
	if last.Valid { d.LastSyncedAt = &last.Time }
	return &d, nil
}

func (r *DomainRepository) GetAnyAvailable(ctx context.Context) (*model.Domain, error) {
	var d model.Domain
	var last sql.NullTime
	err := r.DB.QueryRowContext(ctx, `SELECT id, domain_name, source, sync_status, is_available, last_synced_at, created_at, updated_at
		FROM domains WHERE is_available = 1 AND sync_status = 'active' ORDER BY id ASC LIMIT 1`).
		Scan(&d.ID,&d.DomainName,&d.Source,&d.SyncStatus,&d.IsAvailable,&last,&d.CreatedAt,&d.UpdatedAt)
	if err != nil { return nil, err }
	if last.Valid { d.LastSyncedAt = &last.Time }
	return &d, nil
}

func (r *DomainRepository) Disable(ctx context.Context, id uint64) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE domains SET sync_status='disabled', is_available=0, updated_at=NOW() WHERE id=?`, id)
	return err
}

func (r *DomainRepository) UpsertFromGoDaddy(ctx context.Context, names []string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil { return err }
	defer tx.Rollback()
	seen := map[string]struct{}{}
	for _, n := range names {
		n = strings.ToLower(strings.TrimSpace(n))
		if n == "" { continue }
		seen[n] = struct{}{}
		_, err = tx.ExecContext(ctx, `INSERT INTO domains(domain_name, source, sync_status, is_available, last_synced_at, created_at, updated_at)
			VALUES (?, 'godaddy', 'active', 1, NOW(), NOW(), NOW())
			ON DUPLICATE KEY UPDATE
			last_synced_at = VALUES(last_synced_at),
			updated_at = NOW(),
			is_available = CASE WHEN sync_status='disabled' THEN 0 ELSE 1 END,
			sync_status = CASE WHEN sync_status='disabled' THEN 'disabled' ELSE 'active' END`, n)
		if err != nil { return err }
	}
	if len(seen) > 0 {
		ph := strings.Repeat("?,", len(seen))
		ph = strings.TrimRight(ph, ",")
		args := make([]any,0,len(seen)+1)
		args = append(args, time.Now())
		vals := make([]any,0,len(seen))
		for k := range seen { vals = append(vals,k) }
		args = append(args, vals...)
		_, err = tx.ExecContext(ctx, `UPDATE domains SET sync_status = CASE WHEN sync_status='disabled' THEN 'disabled' ELSE 'missing' END,
			is_available = 0, updated_at = NOW(), last_synced_at = ? WHERE domain_name NOT IN (`+ph+`)`, args...)
		if err != nil { return err }
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE domains SET sync_status = CASE WHEN sync_status='disabled' THEN 'disabled' ELSE 'missing' END, is_available = 0, updated_at = NOW()`)
		if err != nil { return err }
	}
	return tx.Commit()
}
