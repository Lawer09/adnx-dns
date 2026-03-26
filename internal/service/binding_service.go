package service

import (
	"context"
	"database/sql"
	"strings"

	"adnx_dns/internal/errs"
	"adnx_dns/internal/godaddy"
	"adnx_dns/internal/model"
	"adnx_dns/internal/repository"
	"adnx_dns/internal/util"
)

type BindingService struct {
	Domains        *repository.DomainRepository
	Bindings       *repository.BindingRepository
	GoDaddy        *godaddy.Client
	SubdomainChars int
}

type ResolveRequest struct {
	IPv4      string `json:"ipv4"`
	Subdomain string `json:"subdomain"`
	Domain    string `json:"domain"`
	Unique    bool   `json:"unique"`
}

func (s *BindingService) Resolve(ctx context.Context, req ResolveRequest) (map[string]any, error) {
	ip := strings.TrimSpace(req.IPv4)
	if !util.IsValidIPv4(ip) { return nil, errs.New(errs.CodeInvalidIPv4, "invalid ipv4") }

	sub := util.NormalizeSubdomain(req.Subdomain)
	if sub == "" {
		for i := 0; i < 10; i++ {
			candidate := util.RandomLowercase(s.SubdomainChars)
			if err := util.ValidateSubdomain(candidate); err != nil { continue }
			if _, err := s.Bindings.GetActiveByFQDN(ctx, candidate+"."+strings.ToLower(strings.TrimSpace(req.Domain))); err == sql.ErrNoRows || strings.TrimSpace(req.Domain) == "" {
				sub = candidate
				break
			}
		}
		if sub == "" { sub = util.RandomLowercase(s.SubdomainChars) }
	}
	if err := util.ValidateSubdomain(sub); err != nil { return nil, errs.New(errs.CodeInvalidParam, err.Error()) }

	var dom *model.Domain
	var err error
	if strings.TrimSpace(req.Domain) != "" {
		dom, err = s.Domains.GetAvailableByName(ctx, req.Domain)
		if err != nil {
			if err == sql.ErrNoRows { return nil, errs.New(errs.CodeDomainUnavailable, "domain not available") }
			return nil, errs.New(errs.CodeDatabaseError, err.Error())
		}
	} else {
		dom, err = s.Domains.GetAnyAvailable(ctx)
		if err != nil {
			if err == sql.ErrNoRows { return nil, errs.New(errs.CodeNoAvailableDomain, "no available domain") }
			return nil, errs.New(errs.CodeDatabaseError, err.Error())
		}
	}

	fqdn := sub + "." + dom.DomainName
	removedFQDNs := make([]string, 0)

	if existing, err := s.Bindings.GetActiveByFQDN(ctx, fqdn); err == nil {
		if existing.IPv4 == ip {
			return map[string]any{"action": "unchanged", "ipv4": ip, "subdomain": sub, "domain": dom.DomainName, "fqdn": fqdn, "unique": req.Unique}, nil
		}
		if err := s.GoDaddy.UpsertARecord(ctx, dom.DomainName, sub, ip, 600); err != nil {
			return nil, convertProviderErr(err)
		}
		if err := s.Bindings.UpsertActive(ctx, dom.ID, sub, fqdn, ip); err != nil { return nil, errs.New(errs.CodeDatabaseError, err.Error()) }
		if req.Unique {
			removed, err := s.Bindings.ReleaseByIPv4ExceptFQDN(ctx, ip, fqdn)
			if err != nil { return nil, errs.New(errs.CodeDatabaseError, err.Error()) }
			for _, item := range removed {
				_ = s.GoDaddy.DeleteARecord(ctx, item.Domain, item.Subdomain)
				removedFQDNs = append(removedFQDNs, item.FQDN)
			}
		}
		return map[string]any{"action": "updated", "ipv4": ip, "subdomain": sub, "domain": dom.DomainName, "fqdn": fqdn, "unique": req.Unique, "removed_records": removedFQDNs}, nil
	} else if err != sql.ErrNoRows {
		return nil, errs.New(errs.CodeDatabaseError, err.Error())
	}

	if req.Unique {
		removed, err := s.Bindings.ReleaseByIPv4ExceptFQDN(ctx, ip, fqdn)
		if err != nil { return nil, errs.New(errs.CodeDatabaseError, err.Error()) }
		for _, item := range removed {
			if err := s.GoDaddy.DeleteARecord(ctx, item.Domain, item.Subdomain); err != nil {
				return nil, convertProviderErr(err)
			}
			removedFQDNs = append(removedFQDNs, item.FQDN)
		}
	}

	if err := s.GoDaddy.UpsertARecord(ctx, dom.DomainName, sub, ip, 600); err != nil {
		return nil, convertProviderErr(err)
	}
	if err := s.Bindings.UpsertActive(ctx, dom.ID, sub, fqdn, ip); err != nil { return nil, errs.New(errs.CodeDatabaseError, err.Error()) }
	action := "created"
	if req.Unique && len(removedFQDNs) > 0 { action = "replace" }
	return map[string]any{"action": action, "ipv4": ip, "subdomain": sub, "domain": dom.DomainName, "fqdn": fqdn, "unique": req.Unique, "removed_records": removedFQDNs}, nil
}

func (s *BindingService) ListByIP(ctx context.Context, ip string) (map[string]any, error) {
	ip = strings.TrimSpace(ip)
	if !util.IsValidIPv4(ip) { return nil, errs.New(errs.CodeInvalidIPv4, "invalid ipv4") }
	items, err := s.Bindings.ListActiveByIPv4(ctx, ip)
	if err != nil { return nil, errs.New(errs.CodeDatabaseError, err.Error()) }
	if len(items) == 0 { return nil, errs.New(errs.CodeIPNoBindings, "ip has no bound domains") }
	return map[string]any{"ipv4": ip, "records": items}, nil
}

type UnbindRequest struct {
	IPv4 string `json:"ipv4"`
	FQDN string `json:"fqdn"`
}

func (s *BindingService) Unbind(ctx context.Context, req UnbindRequest) (map[string]any, error) {
	ip := strings.TrimSpace(req.IPv4)
	fqdn := strings.ToLower(strings.TrimSpace(req.FQDN))
	if !util.IsValidIPv4(ip) { return nil, errs.New(errs.CodeInvalidIPv4, "invalid ipv4") }
	if fqdn == "" { return nil, errs.New(errs.CodeInvalidParam, "fqdn is required") }
	binding, err := s.Bindings.GetActiveByFQDN(ctx, fqdn)
	if err != nil {
		if err == sql.ErrNoRows { return nil, errs.New(errs.CodeFQDNNotFound, "fqdn not found") }
		return nil, errs.New(errs.CodeDatabaseError, err.Error())
	}
	if binding.IPv4 != ip { return nil, errs.New(errs.CodeIPFQDNMismatch, "fqdn and ip do not match") }
	if err := s.GoDaddy.DeleteARecord(ctx, binding.Domain, binding.Subdomain); err != nil { return nil, convertProviderErr(err) }
	if err := s.Bindings.ReleaseByID(ctx, binding.ID); err != nil { return nil, errs.New(errs.CodeDatabaseError, err.Error()) }
	return map[string]any{"ipv4": ip, "fqdn": fqdn, "action": "deleted"}, nil
}

func convertProviderErr(err error) error {
	if _, ok := err.(*godaddy.ErrRateLimited); ok { return errs.New(errs.CodeRateLimited, err.Error()) }
	return errs.New(errs.CodeProviderError, err.Error())
}
