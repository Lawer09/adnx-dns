package model

import "time"

type Domain struct {
	ID           uint64
	DomainName   string
	Source       string
	SyncStatus   string
	IsAvailable  bool
	LastSyncedAt *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type IPBinding struct {
	ID        uint64
	DomainID  uint64
	Subdomain string
	FQDN      string
	IPv4      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
