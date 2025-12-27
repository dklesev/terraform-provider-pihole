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

func TestClient_GetClients(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/clients":
			json.NewEncoder(w).Encode(ClientsResponse{
				Clients: []PiholeClient{
					{ID: 1, Client: "192.168.1.100", Comment: "Test client", Groups: []int64{0}},
					{ID: 2, Client: "AA:BB:CC:DD:EE:FF", Comment: "MAC client", Groups: []int64{0, 1}},
				},
				Took: 0.001,
			})
		case "/api/clients/192.168.1.100":
			json.NewEncoder(w).Encode(ClientsResponse{
				Clients: []PiholeClient{
					{ID: 1, Client: "192.168.1.100", Comment: "Test client", Groups: []int64{0}},
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

	// Test get all clients
	clients, err := client.GetClients(ctx, "")
	if err != nil {
		t.Fatalf("GetClients() error = %v", err)
	}
	if len(clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(clients))
	}

	// Test get specific client
	clients, err = client.GetClients(ctx, "192.168.1.100")
	if err != nil {
		t.Fatalf("GetClients(192.168.1.100) error = %v", err)
	}
	if len(clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(clients))
	}
	if clients[0].Client != "192.168.1.100" {
		t.Errorf("Expected client '192.168.1.100', got %q", clients[0].Client)
	}
}

func TestClient_CreateClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/clients":
			if r.Method == http.MethodPost {
				var req map[string]interface{}
				json.NewDecoder(r.Body).Decode(&req)

				clientID, _ := req["client"].(string)
				comment, _ := req["comment"].(string)

				json.NewEncoder(w).Encode(ClientsResponse{
					Clients: []PiholeClient{
						{
							ID:        1,
							Client:    clientID,
							Comment:   comment,
							Groups:    []int64{0},
							DateAdded: 1234567890,
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
	piholeClient := &PiholeClient{
		Client:  "192.168.1.200",
		Comment: "New test client",
		Groups:  []int64{0},
	}

	created, err := client.CreateClient(ctx, piholeClient)
	if err != nil {
		t.Fatalf("CreateClient() error = %v", err)
	}
	if created.Client != "192.168.1.200" {
		t.Errorf("Expected client '192.168.1.200', got %q", created.Client)
	}
	if created.ID != 1 {
		t.Errorf("Expected ID 1, got %d", created.ID)
	}
}

func TestClient_UpdateClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/clients/192.168.1.100":
			if r.Method == http.MethodPut {
				var req map[string]interface{}
				json.NewDecoder(r.Body).Decode(&req)

				clientID, _ := req["client"].(string)
				comment, _ := req["comment"].(string)

				json.NewEncoder(w).Encode(ClientsResponse{
					Clients: []PiholeClient{
						{
							ID:           1,
							Client:       clientID,
							Comment:      comment,
							Groups:       []int64{0},
							DateModified: 1234567891,
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
	piholeClient := &PiholeClient{
		Client:  "192.168.1.100",
		Comment: "Updated comment",
		Groups:  []int64{0},
	}

	updated, err := client.UpdateClient(ctx, "192.168.1.100", piholeClient)
	if err != nil {
		t.Fatalf("UpdateClient() error = %v", err)
	}
	if updated.Comment != "Updated comment" {
		t.Errorf("Expected comment 'Updated comment', got %q", updated.Comment)
	}
}

func TestClient_DeleteClient(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/clients/192.168.1.100":
			if r.Method == http.MethodDelete {
				deleted = true
				w.WriteHeader(http.StatusNoContent)
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
	err = client.DeleteClient(ctx, "192.168.1.100")
	if err != nil {
		t.Fatalf("DeleteClient() error = %v", err)
	}
	if !deleted {
		t.Error("Expected DELETE request to be made")
	}
}
