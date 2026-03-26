package model

import "time"

type Domain struct {
	ID           uint64     `json:"id"`
	DomainName   string     `json:"domain"`
	Source       string     `json:"source"`
	SyncStatus   string     `json:"sync_status"`
	IsAvailable  bool       `json:"enabled"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type DomainDetail struct {
	Domain
	Records []IPBinding `json:"records"`
}

type IPBinding struct {
	ID        uint64    `json:"id"`
	DomainID  uint64    `json:"domain_id"`
	Domain    string    `json:"domain,omitempty"`
	Subdomain string    `json:"subdomain"`
	FQDN      string    `json:"fqdn"`
	IPv4      string    `json:"ipv4"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
