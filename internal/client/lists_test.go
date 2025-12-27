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

func TestClient_GetLists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/lists":
			listType := r.URL.Query().Get("type")
			lists := []List{
				{ID: 1, Address: "https://example.com/blocklist.txt", Type: "block", Enabled: true, Groups: []int64{0}},
				{ID: 2, Address: "https://example.com/allowlist.txt", Type: "allow", Enabled: true, Groups: []int64{0}},
			}
			if listType != "" {
				filtered := []List{}
				for _, l := range lists {
					if l.Type == listType {
						filtered = append(filtered, l)
					}
				}
				lists = filtered
			}
			json.NewEncoder(w).Encode(ListsResponse{
				Lists: lists,
				Took:  0.001,
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

	// Test get all lists
	lists, err := client.GetLists(ctx, "", "")
	if err != nil {
		t.Fatalf("GetLists() error = %v", err)
	}
	if len(lists) != 2 {
		t.Errorf("Expected 2 lists, got %d", len(lists))
	}

	// Test filter by type
	lists, err = client.GetLists(ctx, "block", "")
	if err != nil {
		t.Fatalf("GetLists(block) error = %v", err)
	}
	if len(lists) != 1 {
		t.Errorf("Expected 1 blocklist, got %d", len(lists))
	}
	if lists[0].Type != "block" {
		t.Errorf("Expected type 'block', got %q", lists[0].Type)
	}
}

func TestClient_CreateList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/lists":
			if r.Method == http.MethodPost {
				var req map[string]interface{}
				json.NewDecoder(r.Body).Decode(&req)

				address, _ := req["address"].(string)
				enabled, _ := req["enabled"].(bool)

				json.NewEncoder(w).Encode(ListsResponse{
					Lists: []List{
						{
							ID:        1,
							Address:   address,
							Type:      "block",
							Enabled:   enabled,
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
	list := &List{
		Address: "https://example.com/newlist.txt",
		Type:    "block",
		Enabled: true,
		Groups:  []int64{0},
	}

	created, err := client.CreateList(ctx, list)
	if err != nil {
		t.Fatalf("CreateList() error = %v", err)
	}
	if created.Address != "https://example.com/newlist.txt" {
		t.Errorf("Expected address 'https://example.com/newlist.txt', got %q", created.Address)
	}
	if created.ID != 1 {
		t.Errorf("Expected ID 1, got %d", created.ID)
	}
}

func TestClient_CreateList_ValidationErrors(t *testing.T) {
	client, err := New(Config{URL: "http://localhost", Password: "test"})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Missing type
	_, err = client.CreateList(ctx, &List{
		Address: "https://example.com/list.txt",
	})
	if err == nil {
		t.Error("Expected error for missing type")
	}
}

func TestClient_UpdateList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/lists/list.txt":
			if r.Method == http.MethodPut {
				var req map[string]interface{}
				json.NewDecoder(r.Body).Decode(&req)

				address, _ := req["address"].(string)
				enabled, _ := req["enabled"].(bool)
				comment, _ := req["comment"].(string)

				json.NewEncoder(w).Encode(ListsResponse{
					Lists: []List{
						{
							ID:           1,
							Address:      address,
							Type:         "block",
							Enabled:      enabled,
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
	list := &List{
		Address: "list.txt",
		Type:    "block",
		Enabled: false,
		Comment: "Updated comment",
		Groups:  []int64{0},
	}

	updated, err := client.UpdateList(ctx, "block", "list.txt", list)
	if err != nil {
		t.Fatalf("UpdateList() error = %v", err)
	}
	if updated.Comment != "Updated comment" {
		t.Errorf("Expected comment 'Updated comment', got %q", updated.Comment)
	}
	if updated.Enabled != false {
		t.Errorf("Expected enabled=false, got %v", updated.Enabled)
	}
}

func TestClient_DeleteList(t *testing.T) {
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
		case "/api/lists/list.txt":
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
	err = client.DeleteList(ctx, "block", "list.txt")
	if err != nil {
		t.Fatalf("DeleteList() error = %v", err)
	}
	if !deleted {
		t.Error("Expected DELETE request to be made")
	}
}
