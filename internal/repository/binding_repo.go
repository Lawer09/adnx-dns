package repository

import (
	"context"
	"database/sql"

	"adnx_dns/internal/model"
)

type BindingRepository struct{ DB *sql.DB }

func (r *BindingRepository) GetActiveByFQDN(ctx context.Context, fqdn string) (*model.IPBinding, error) {
	var b model.IPBinding
	err := r.DB.QueryRowContext(ctx, `SELECT id, domain_id, subdomain, fqdn, ipv4, status, created_at, updated_at FROM ip_bindings WHERE fqdn=? AND status='active' LIMIT 1`, fqdn).
		Scan(&b.ID,&b.DomainID,&b.Subdomain,&b.FQDN,&b.IPv4,&b.Status,&b.CreatedAt,&b.UpdatedAt)
	if err != nil { return nil, err }
	return &b, nil
}

func (r *BindingRepository) GetActiveByIPv4(ctx context.Context, ip string) (*model.IPBinding, error) {
	var b model.IPBinding
	err := r.DB.QueryRowContext(ctx, `SELECT id, domain_id, subdomain, fqdn, ipv4, status, created_at, updated_at FROM ip_bindings WHERE ipv4=? AND status='active' LIMIT 1`, ip).
		Scan(&b.ID,&b.DomainID,&b.Subdomain,&b.FQDN,&b.IPv4,&b.Status,&b.CreatedAt,&b.UpdatedAt)
	if err != nil { return nil, err }
	return &b, nil
}

func (r *BindingRepository) UpsertActive(ctx context.Context, domainID uint64, subdomain, fqdn, ip string) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO ip_bindings(domain_id, subdomain, fqdn, ipv4, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'active', NOW(), NOW())
		ON DUPLICATE KEY UPDATE domain_id=VALUES(domain_id), subdomain=VALUES(subdomain), ipv4=VALUES(ipv4), status='active', updated_at=NOW()`, domainID, subdomain, fqdn, ip)
	return err
}

func (r *BindingRepository) ReleaseByFQDNOrIP(ctx context.Context, subdomain, domain, ip string) (*model.IPBinding, error) {
	query := `SELECT id, domain_id, subdomain, fqdn, ipv4, status, created_at, updated_at FROM ip_bindings WHERE status='active' AND `
	args := []any{}
	if ip != "" {
		query += `ipv4=? LIMIT 1`
		args = append(args, ip)
	} else {
		fqdn := subdomain + "." + domain
		query += `fqdn=? LIMIT 1`
		args = append(args, fqdn)
	}
	var b model.IPBinding
	err := r.DB.QueryRowContext(ctx, query, args...).Scan(&b.ID,&b.DomainID,&b.Subdomain,&b.FQDN,&b.IPv4,&b.Status,&b.CreatedAt,&b.UpdatedAt)
	if err != nil { return nil, err }
	_, err = r.DB.ExecContext(ctx, `UPDATE ip_bindings SET status='released', updated_at=NOW() WHERE id=?`, b.ID)
	if err != nil { return nil, err }
	return &b, nil
}
