package repository

import (
	"context"
	"database/sql"
	"strings"

	"adnx_dns/internal/model"
)

type BindingRepository struct{ DB *sql.DB }

func scanBinding(scanner interface{ Scan(dest ...any) error }) (*model.IPBinding, error) {
	var b model.IPBinding
	if err := scanner.Scan(&b.ID, &b.DomainID, &b.Domain, &b.Subdomain, &b.FQDN, &b.IPv4, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BindingRepository) ListByDomainID(ctx context.Context, domainID uint64, activeOnly bool) ([]model.IPBinding, error) {
	query := `SELECT b.id, b.domain_id, d.domain_name, b.subdomain, b.fqdn, b.ipv4, b.status, b.created_at, b.updated_at
		FROM ip_bindings b JOIN domains d ON d.id=b.domain_id WHERE b.domain_id=?`
	if activeOnly { query += ` AND b.status='active'` }
	query += ` ORDER BY b.subdomain ASC`
	rows, err := r.DB.QueryContext(ctx, query, domainID)
	if err != nil { return nil, err }
	defer rows.Close()
	out := make([]model.IPBinding, 0)
	for rows.Next() {
		b, err := scanBinding(rows)
		if err != nil { return nil, err }
		out = append(out, *b)
	}
	return out, rows.Err()
}

func (r *BindingRepository) GetActiveByFQDN(ctx context.Context, fqdn string) (*model.IPBinding, error) {
	row := r.DB.QueryRowContext(ctx, `SELECT b.id, b.domain_id, d.domain_name, b.subdomain, b.fqdn, b.ipv4, b.status, b.created_at, b.updated_at
		FROM ip_bindings b JOIN domains d ON d.id=b.domain_id WHERE b.fqdn=? AND b.status='active' LIMIT 1`, strings.ToLower(strings.TrimSpace(fqdn)))
	return scanBinding(row)
}

func (r *BindingRepository) ListActiveByIPv4(ctx context.Context, ip string) ([]model.IPBinding, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT b.id, b.domain_id, d.domain_name, b.subdomain, b.fqdn, b.ipv4, b.status, b.created_at, b.updated_at
		FROM ip_bindings b JOIN domains d ON d.id=b.domain_id WHERE b.ipv4=? AND b.status='active' ORDER BY b.fqdn ASC`, strings.TrimSpace(ip))
	if err != nil { return nil, err }
	defer rows.Close()
	out := make([]model.IPBinding, 0)
	for rows.Next() {
		b, err := scanBinding(rows)
		if err != nil { return nil, err }
		out = append(out, *b)
	}
	return out, rows.Err()
}

func (r *BindingRepository) UpsertActive(ctx context.Context, domainID uint64, subdomain, fqdn, ip string) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO ip_bindings(domain_id, subdomain, fqdn, ipv4, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'active', NOW(), NOW())
		ON DUPLICATE KEY UPDATE domain_id=VALUES(domain_id), subdomain=VALUES(subdomain), ipv4=VALUES(ipv4), status='active', updated_at=NOW()`, domainID, subdomain, fqdn, ip)
	return err
}

func (r *BindingRepository) ReleaseByID(ctx context.Context, id uint64) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE ip_bindings SET status='released', updated_at=NOW() WHERE id=?`, id)
	return err
}

func (r *BindingRepository) ReleaseByIPv4ExceptFQDN(ctx context.Context, ip, keepFQDN string) ([]model.IPBinding, error) {
	rows, err := r.ListActiveByIPv4(ctx, ip)
	if err != nil { return nil, err }
	removed := make([]model.IPBinding, 0)
	for _, item := range rows {
		if strings.EqualFold(item.FQDN, keepFQDN) { continue }
		if err := r.ReleaseByID(ctx, item.ID); err != nil { return nil, err }
		removed = append(removed, item)
	}
	return removed, nil
}
