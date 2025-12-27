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

func TestClient_GetGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/groups":
			json.NewEncoder(w).Encode(GroupsResponse{
				Groups: []Group{
					{ID: 0, Name: "Default", Enabled: true},
					{ID: 1, Name: "Custom", Enabled: true, Description: "Custom group"},
				},
				Took: 0.001,
			})
		case "/api/groups/Custom":
			json.NewEncoder(w).Encode(GroupsResponse{
				Groups: []Group{
					{ID: 1, Name: "Custom", Enabled: true, Description: "Custom group"},
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

	// Test get all groups
	groups, err := client.GetGroups(ctx, "")
	if err != nil {
		t.Fatalf("GetGroups() error = %v", err)
	}
	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}

	// Test get specific group
	groups, err = client.GetGroups(ctx, "Custom")
	if err != nil {
		t.Fatalf("GetGroups(Custom) error = %v", err)
	}
	if len(groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups))
	}
	if groups[0].Name != "Custom" {
		t.Errorf("Expected group name 'Custom', got %q", groups[0].Name)
	}
}

func TestClient_CreateGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/groups":
			if r.Method == http.MethodPost {
				var req map[string]interface{}
				json.NewDecoder(r.Body).Decode(&req)

				name, _ := req["name"].(string)
				enabled, _ := req["enabled"].(bool)
				comment, _ := req["comment"].(string)

				json.NewEncoder(w).Encode(GroupsResponse{
					Groups: []Group{
						{
							ID:          1,
							Name:        name,
							Enabled:     enabled,
							Description: comment,
							DateAdded:   1234567890,
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
	group := &Group{
		Name:        "TestGroup",
		Enabled:     true,
		Description: "Test description",
	}

	created, err := client.CreateGroup(ctx, group)
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	if created.Name != "TestGroup" {
		t.Errorf("Expected name 'TestGroup', got %q", created.Name)
	}
	if created.ID != 1 {
		t.Errorf("Expected ID 1, got %d", created.ID)
	}
}

func TestClient_UpdateGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"session": map[string]interface{}{
					"valid": true,
					"sid":   "test-sid",
				},
			})
		case "/api/groups/OldName":
			if r.Method == http.MethodPut {
				var req map[string]interface{}
				json.NewDecoder(r.Body).Decode(&req)

				name, _ := req["name"].(string)
				enabled, _ := req["enabled"].(bool)

				json.NewEncoder(w).Encode(GroupsResponse{
					Groups: []Group{
						{
							ID:           1,
							Name:         name,
							Enabled:      enabled,
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
	group := &Group{
		Name:    "NewName",
		Enabled: false,
	}

	updated, err := client.UpdateGroup(ctx, "OldName", group)
	if err != nil {
		t.Fatalf("UpdateGroup() error = %v", err)
	}
	if updated.Name != "NewName" {
		t.Errorf("Expected name 'NewName', got %q", updated.Name)
	}
	if updated.Enabled != false {
		t.Errorf("Expected enabled=false, got %v", updated.Enabled)
	}
}

func TestClient_DeleteGroup(t *testing.T) {
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
		case "/api/groups/TestGroup":
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
	err = client.DeleteGroup(ctx, "TestGroup")
	if err != nil {
		t.Fatalf("DeleteGroup() error = %v", err)
	}
	if !deleted {
		t.Error("Expected DELETE request to be made")
	}
}
