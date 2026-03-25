package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"adnx_dns/internal/godaddy"
	"adnx_dns/internal/repository"
	"adnx_dns/internal/util"
)

type BindingService struct {
	Domains        *repository.DomainRepository
	Bindings       *repository.BindingRepository
	GoDaddy        *godaddy.Client
	SubdomainChars int
}

type BindRequest struct {
	IPv4      string `json:"ipv4"`
	Subdomain string `json:"subdomain"`
	Domain    string `json:"domain"`
}

type BindResponse struct {
	Message   string `json:"message"`
	Domain    string `json:"domain"`
	Subdomain string `json:"subdomain"`
	FQDN      string `json:"fqdn"`
	IPv4      string `json:"ipv4"`
	Changed   bool   `json:"changed"`
}

func (s *BindingService) Bind(ctx context.Context, req BindRequest) (*BindResponse, int, error) {
	ip := strings.TrimSpace(req.IPv4)
	if !util.IsValidIPv4(ip) {
		return nil, 400, errors.New("invalid ipv4")
	}
	sub := util.NormalizeSubdomain(req.Subdomain)
	if sub == "" {
		sub = util.RandomLowercase(s.SubdomainChars)
	}
	if err := util.ValidateSubdomain(sub); err != nil {
		return nil, 400, err
	}

	var domainName string
	var domainID uint64
	if strings.TrimSpace(req.Domain) != "" {
		d, err := s.Domains.GetAvailableByName(ctx, req.Domain)
		if err != nil {
			if err == sql.ErrNoRows { return nil, 404, errors.New("domain not available") }
			return nil, 500, err
		}
		domainName, domainID = d.DomainName, d.ID
	} else {
		d, err := s.Domains.GetAnyAvailable(ctx)
		if err != nil {
			if err == sql.ErrNoRows { return nil, 409, errors.New("no available domain") }
			return nil, 500, err
		}
		domainName, domainID = d.DomainName, d.ID
	}

	fqdn := sub + "." + domainName
	if existingByFQDN, err := s.Bindings.GetActiveByFQDN(ctx, fqdn); err == nil {
		if existingByFQDN.IPv4 == ip {
			return &BindResponse{Message:"already bound", Domain:domainName, Subdomain:sub, FQDN:fqdn, IPv4:ip, Changed:false}, 200, nil
		}
		if err := s.GoDaddy.UpsertARecord(ctx, domainName, sub, ip, 600); err != nil {
			if _, ok := err.(*godaddy.ErrRateLimited); ok { return nil, 429, err }
			return nil, 502, err
		}
		if err := s.Bindings.UpsertActive(ctx, domainID, sub, fqdn, ip); err != nil { return nil, 500, err }
		return &BindResponse{Message:"fqdn existed, switched to new ip", Domain:domainName, Subdomain:sub, FQDN:fqdn, IPv4:ip, Changed:true}, 200, nil
	} else if err != sql.ErrNoRows {
		return nil, 500, err
	}

	if existingByIP, err := s.Bindings.GetActiveByIPv4(ctx, ip); err == nil {
		parts := strings.SplitN(existingByIP.FQDN, ".", 2)
		oldSub := existingByIP.Subdomain
		oldDomain := ""
		if len(parts) == 2 { oldDomain = parts[1] }
		if oldDomain != "" {
			if err := s.GoDaddy.DeleteARecord(ctx, oldDomain, oldSub); err != nil {
				if _, ok := err.(*godaddy.ErrRateLimited); ok { return nil, 429, err }
				return nil, 502, err
			}
		}
		_, _ = s.Bindings.ReleaseByFQDNOrIP(ctx, "", "", ip)
	} else if err != sql.ErrNoRows {
		return nil, 500, err
	}

	if err := s.GoDaddy.UpsertARecord(ctx, domainName, sub, ip, 600); err != nil {
		if _, ok := err.(*godaddy.ErrRateLimited); ok { return nil, 429, err }
		return nil, 502, err
	}
	if err := s.Bindings.UpsertActive(ctx, domainID, sub, fqdn, ip); err != nil { return nil, 500, err }
	return &BindResponse{Message:"bind success", Domain:domainName, Subdomain:sub, FQDN:fqdn, IPv4:ip, Changed:true}, 200, nil
}

type UnbindRequest struct {
	Subdomain string `json:"subdomain"`
	Domain    string `json:"domain"`
	IPv4      string `json:"ipv4"`
}

func (s *BindingService) Unbind(ctx context.Context, req UnbindRequest) (map[string]any, int, error) {
	if strings.TrimSpace(req.IPv4) == "" && (strings.TrimSpace(req.Subdomain) == "" || strings.TrimSpace(req.Domain) == "") {
		return nil, 400, errors.New("provide ipv4 or subdomain+domain")
	}
	b, err := s.Bindings.ReleaseByFQDNOrIP(ctx, util.NormalizeSubdomain(req.Subdomain), strings.ToLower(strings.TrimSpace(req.Domain)), strings.TrimSpace(req.IPv4))
	if err != nil {
		if err == sql.ErrNoRows { return nil, 404, errors.New("binding not found") }
		return nil, 500, err
	}
	parts := strings.SplitN(b.FQDN, ".", 2)
	if len(parts) != 2 { return nil, 500, fmt.Errorf("invalid fqdn stored: %s", b.FQDN) }
	if err := s.GoDaddy.DeleteARecord(ctx, parts[1], parts[0]); err != nil {
		if _, ok := err.(*godaddy.ErrRateLimited); ok { return nil, 429, err }
		return nil, 502, err
	}
	return map[string]any{"message":"unbind success", "fqdn":b.FQDN, "ipv4":b.IPv4}, 200, nil
}
