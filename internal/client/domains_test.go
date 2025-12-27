// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetDomains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/domains":
			json.NewEncoder(w).Encode(DomainsResponse{
				Domains: []Domain{
					{ID: 1, Domain: "ads.example.com", Type: "deny", Kind: "exact", Enabled: true},
					{ID: 2, Domain: "^ads\\..*", Type: "deny", Kind: "regex", Enabled: true},
				},
				Took: 0.001,
			})
		case "/api/domains/deny":
			json.NewEncoder(w).Encode(DomainsResponse{
				Domains: []Domain{
					{ID: 1, Domain: "ads.example.com", Type: "deny", Kind: "exact", Enabled: true},
					{ID: 2, Domain: "^ads\\..*", Type: "deny", Kind: "regex", Enabled: true},
				},
				Took: 0.001,
			})
		case "/api/domains/deny/exact":
			json.NewEncoder(w).Encode(DomainsResponse{
				Domains: []Domain{
					{ID: 1, Domain: "ads.example.com", Type: "deny", Kind: "exact", Enabled: true},
				},
				Took: 0.001,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := New(Config{URL: server.URL, Password: "test"})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test get all domains
	domains, err := client.GetDomains(ctx, "", "", "")
	if err != nil {
		t.Fatalf("GetDomains() error = %v", err)
	}
	if len(domains) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(domains))
	}

	// Test filter by type
	domains, err = client.GetDomains(ctx, "deny", "", "")
	if err != nil {
		t.Fatalf("GetDomains(deny) error = %v", err)
	}
	if len(domains) != 2 {
		t.Errorf("Expected 2 deny domains, got %d", len(domains))
	}

	// Test filter by type and kind
	domains, err = client.GetDomains(ctx, "deny", "exact", "")
	if err != nil {
		t.Fatalf("GetDomains(deny, exact) error = %v", err)
	}
	if len(domains) != 1 {
		t.Errorf("Expected 1 exact deny domain, got %d", len(domains))
	}
}

func TestClient_CreateDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/domains/deny/exact":
			if r.Method == http.MethodPost {
				var req map[string]interface{}
				json.NewDecoder(r.Body).Decode(&req)

				domain, _ := req["domain"].(string)
				enabled, _ := req["enabled"].(bool)

				json.NewEncoder(w).Encode(DomainsResponse{
					Domains: []Domain{
						{
							ID:      1,
							Domain:  domain,
							Type:    "deny",
							Kind:    "exact",
							Enabled: enabled,
						},
					},
					Took: 0.001,
				})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := New(Config{URL: server.URL, Password: "test"})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	domain := &Domain{
		Domain:  "test.example.com",
		Type:    "deny",
		Kind:    "exact",
		Enabled: true,
	}

	created, err := client.CreateDomain(ctx, domain)
	if err != nil {
		t.Fatalf("CreateDomain() error = %v", err)
	}
	if created.Domain != "test.example.com" {
		t.Errorf("Expected domain 'test.example.com', got %q", created.Domain)
	}
	if created.Type != "deny" {
		t.Errorf("Expected type 'deny', got %q", created.Type)
	}
}

func TestClient_CreateDomain_ValidationErrors(t *testing.T) {
	client, err := New(Config{URL: "http://localhost", Password: "test"})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Missing type
	_, err = client.CreateDomain(ctx, &Domain{
		Domain: "test.com",
		Kind:   "exact",
	})
	if err == nil {
		t.Error("Expected error for missing type")
	}

	// Missing kind
	_, err = client.CreateDomain(ctx, &Domain{
		Domain: "test.com",
		Type:   "deny",
	})
	if err == nil {
		t.Error("Expected error for missing kind")
	}
}
