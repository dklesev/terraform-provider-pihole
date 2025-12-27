// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				URL:      "http://pi.hole",
				Password: "test",
			},
			wantErr: false,
		},
		{
			name: "valid config with port",
			cfg: Config{
				URL:      "http://pi.hole:8080",
				Password: "test",
			},
			wantErr: false,
		},
		{
			name: "valid config with path",
			cfg: Config{
				URL:      "http://pi.hole/admin",
				Password: "test",
			},
			wantErr: false,
		},
		{
			name: "valid config already has /api",
			cfg: Config{
				URL:      "http://pi.hole/api",
				Password: "test",
			},
			wantErr: false,
		},
		{
			name: "empty URL",
			cfg: Config{
				URL:      "",
				Password: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			cfg: Config{
				URL:      "://invalid",
				Password: "test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("New() returned nil client without error")
			}
		})
	}
}

func TestClient_Authenticate(t *testing.T) {
	tests := []struct {
		name           string
		password       string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
	}{
		{
			name:     "successful auth",
			password: "test123",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/auth" {
					if r.Method == http.MethodGet {
						// Check auth status - return unauthenticated
						json.NewEncoder(w).Encode(map[string]interface{}{
							"session": map[string]interface{}{
								"valid": false,
							},
						})
					} else if r.Method == http.MethodPost {
						// Login
						json.NewEncoder(w).Encode(map[string]interface{}{
							"session": map[string]interface{}{
								"valid":    true,
								"sid":      "test-session-id",
								"validity": 300,
							},
						})
					}
				}
			},
			wantErr: false,
		},
		{
			name:     "no password required",
			password: "",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/auth" && r.Method == http.MethodGet {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"session": map[string]interface{}{
							"valid": true,
							"sid":   "no-auth-session",
						},
					})
				}
			},
			wantErr: false,
		},
		{
			name:     "wrong password",
			password: "wrongpassword",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/auth" {
					if r.Method == http.MethodGet {
						json.NewEncoder(w).Encode(map[string]interface{}{
							"session": map[string]interface{}{
								"valid": false,
							},
						})
					} else if r.Method == http.MethodPost {
						w.WriteHeader(http.StatusUnauthorized)
						json.NewEncoder(w).Encode(map[string]interface{}{
							"error": map[string]interface{}{
								"key":     "unauthorized",
								"message": "Invalid password",
							},
						})
					}
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			client, err := New(Config{
				URL:      server.URL,
				Password: tt.password,
			})
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			ctx := context.Background()
			err = client.Authenticate(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Authenticate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_SessionRefresh(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth" {
			if r.Method == http.MethodGet {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"session": map[string]interface{}{
						"valid": false,
					},
				})
			} else if r.Method == http.MethodPost {
				callCount++
				json.NewEncoder(w).Encode(map[string]interface{}{
					"session": map[string]interface{}{
						"valid":    true,
						"sid":      "session-" + string(rune('0'+callCount)),
						"validity": 1, // 1 second validity
					},
				})
			}
		}
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Password: "test",
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// First auth
	err = client.Authenticate(ctx)
	if err != nil {
		t.Fatalf("First auth failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 auth call, got %d", callCount)
	}

	// Wait for session to expire
	time.Sleep(2 * time.Second)

	// Second auth should trigger refresh
	err = client.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Second auth failed: %v", err)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 auth calls after expiry, got %d", callCount)
	}
}

func TestClient_Request_Errors(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		errContains    string
	}{
		{
			name: "400 bad request",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/auth" {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"session": map[string]interface{}{
							"valid": true,
							"sid":   "test-sid",
						},
					})
				} else {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": map[string]interface{}{
							"key":     "bad_request",
							"message": "Invalid input",
						},
					})
				}
			},
			wantErr:     true,
			errContains: "Invalid input",
		},
		{
			name: "404 not found",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/auth" {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"session": map[string]interface{}{
							"valid": true,
							"sid":   "test-sid",
						},
					})
				} else {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": map[string]interface{}{
							"key":     "not_found",
							"message": "Resource not found",
						},
					})
				}
			},
			wantErr:     true,
			errContains: "not_found",
		},
		{
			name: "500 server error",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/auth" {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"session": map[string]interface{}{
							"valid": true,
							"sid":   "test-sid",
						},
					})
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": map[string]interface{}{
							"key":     "internal_error",
							"message": "Something went wrong",
						},
					})
				}
			},
			wantErr: true,
			// Note: With retry logic enabled, 5xx errors trigger retries and the error message
			// reflects retry exhaustion rather than the response body content
			errContains: "giving up",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			// Disable retries for error tests to get consistent error messages
			client, err := New(Config{
				URL:      server.URL,
				Password: "test",
				RetryMax: -1, // Explicitly disable retries
			})
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			ctx := context.Background()
			_, err = client.Get(ctx, "groups")
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got %v", tt.errContains, err)
				}
			}
		})
	}
}
