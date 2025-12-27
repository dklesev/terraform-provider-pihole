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

func TestClient_GetDNSBlocking(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/dns/blocking":
			if r.Method == http.MethodGet {
				json.NewEncoder(w).Encode(DNSBlockingResponse{
					Blocking: "enabled",
					Timer:    nil,
					Took:     0.001,
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
	blocking, err := client.GetDNSBlocking(ctx)
	if err != nil {
		t.Fatalf("GetDNSBlocking() error = %v", err)
	}
	if blocking.Blocking != "enabled" {
		t.Errorf("Expected blocking 'enabled', got %q", blocking.Blocking)
	}
	if blocking.Timer != nil {
		t.Errorf("Expected nil timer, got %v", *blocking.Timer)
	}
}

func TestClient_GetDNSBlocking_WithTimer(t *testing.T) {
	timer := 300.0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/dns/blocking":
			if r.Method == http.MethodGet {
				json.NewEncoder(w).Encode(DNSBlockingResponse{
					Blocking: "disabled",
					Timer:    &timer,
					Took:     0.001,
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
	blocking, err := client.GetDNSBlocking(ctx)
	if err != nil {
		t.Fatalf("GetDNSBlocking() error = %v", err)
	}
	if blocking.Blocking != "disabled" {
		t.Errorf("Expected blocking 'disabled', got %q", blocking.Blocking)
	}
	if blocking.Timer == nil {
		t.Error("Expected timer to be set")
	} else if *blocking.Timer != 300.0 {
		t.Errorf("Expected timer 300.0, got %v", *blocking.Timer)
	}
}

func TestClient_SetDNSBlocking_Enable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/dns/blocking":
			if r.Method == http.MethodPost {
				var req DNSBlockingRequest
				json.NewDecoder(r.Body).Decode(&req)

				status := "disabled"
				if req.Blocking {
					status = "enabled"
				}

				json.NewEncoder(w).Encode(DNSBlockingResponse{
					Blocking: status,
					Timer:    req.Timer,
					Took:     0.001,
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
	blocking, err := client.SetDNSBlocking(ctx, true, nil)
	if err != nil {
		t.Fatalf("SetDNSBlocking() error = %v", err)
	}
	if blocking.Blocking != "enabled" {
		t.Errorf("Expected blocking 'enabled', got %q", blocking.Blocking)
	}
}

func TestClient_SetDNSBlocking_DisableWithTimer(t *testing.T) {
	timer := 600.0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/dns/blocking":
			if r.Method == http.MethodPost {
				var req DNSBlockingRequest
				json.NewDecoder(r.Body).Decode(&req)

				json.NewEncoder(w).Encode(DNSBlockingResponse{
					Blocking: "disabled",
					Timer:    req.Timer,
					Took:     0.001,
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
	blocking, err := client.SetDNSBlocking(ctx, false, &timer)
	if err != nil {
		t.Fatalf("SetDNSBlocking() error = %v", err)
	}
	if blocking.Blocking != "disabled" {
		t.Errorf("Expected blocking 'disabled', got %q", blocking.Blocking)
	}
	if blocking.Timer == nil {
		t.Error("Expected timer to be set")
	} else if *blocking.Timer != 600.0 {
		t.Errorf("Expected timer 600.0, got %v", *blocking.Timer)
	}
}
