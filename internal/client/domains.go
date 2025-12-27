// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// GetDomains retrieves domains with optional filters.
func (c *Client) GetDomains(ctx context.Context, domainType, kind, domain string) ([]Domain, error) {
	segments := []string{"domains"}
	if domainType != "" {
		segments = append(segments, domainType)
	}
	if kind != "" {
		segments = append(segments, kind)
	}
	if domain != "" {
		segments = append(segments, url.PathEscape(domain))
	}
	path := strings.Join(segments, "/")

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result DomainsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse domains response: %w", err)
	}

	return result.Domains, nil
}

// GetDomain retrieves a specific domain.
func (c *Client) GetDomain(ctx context.Context, domainType, kind, domain string) (*Domain, error) {
	domains, err := c.GetDomains(ctx, domainType, kind, domain)
	if err != nil {
		return nil, err
	}

	// Find exact match
	for _, d := range domains {
		if d.Domain == domain && d.Type == domainType && d.Kind == kind {
			return &d, nil
		}
	}

	return nil, nil // Not found
}

// CreateDomain creates a new domain entry.
func (c *Client) CreateDomain(ctx context.Context, domain *Domain) (*Domain, error) {
	if domain.Type == "" || domain.Kind == "" {
		return nil, fmt.Errorf("domain type and kind are required")
	}

	payload := map[string]interface{}{
		"domain":  domain.Domain,
		"enabled": domain.Enabled,
	}
	if domain.Comment != "" {
		payload["comment"] = domain.Comment
	}
	if len(domain.Groups) > 0 {
		payload["groups"] = domain.Groups
	}

	path := fmt.Sprintf("domains/%s/%s", domain.Type, domain.Kind)
	resp, err := c.Post(ctx, path, payload)
	if err != nil {
		return nil, err
	}

	var result DomainsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse create domain response: %w", err)
	}

	if len(result.Domains) == 0 {
		return nil, fmt.Errorf("no domain returned in response")
	}

	return &result.Domains[0], nil
}

// UpdateDomain updates an existing domain entry.
func (c *Client) UpdateDomain(ctx context.Context, originalType, originalKind, originalDomain string, domain *Domain) (*Domain, error) {
	payload := map[string]interface{}{
		"domain":  domain.Domain,
		"enabled": domain.Enabled,
		"comment": domain.Comment,
		"groups":  domain.Groups,
	}

	// If type/kind changed, we need to specify the new values
	if domain.Type != originalType || domain.Kind != originalKind {
		payload["type"] = domain.Type
		payload["kind"] = domain.Kind
	}

	path := fmt.Sprintf("domains/%s/%s/%s", originalType, originalKind, url.PathEscape(originalDomain))
	resp, err := c.Put(ctx, path, payload)
	if err != nil {
		return nil, err
	}

	var result DomainsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse update domain response: %w", err)
	}

	// If domains array is empty (may happen on domain/type/kind change), fetch by new values
	if len(result.Domains) == 0 {
		updatedDomain, err := c.GetDomain(ctx, domain.Type, domain.Kind, domain.Domain)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch updated domain: %w", err)
		}
		if updatedDomain == nil {
			return nil, fmt.Errorf("no domain returned in response and updated domain not found")
		}
		return updatedDomain, nil
	}

	return &result.Domains[0], nil
}

// DeleteDomain deletes a domain entry.
func (c *Client) DeleteDomain(ctx context.Context, domainType, kind, domain string) error {
	path := fmt.Sprintf("domains/%s/%s/%s", domainType, kind, url.PathEscape(domain))
	_, err := c.Delete(ctx, path)
	return err
}
