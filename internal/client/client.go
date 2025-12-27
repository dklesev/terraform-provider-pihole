// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

// Package client provides a Go client for the Pi-hole FTL API v6.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	// DefaultTimeout is the default HTTP timeout for API requests.
	DefaultTimeout = 30 * time.Second

	// SessionRefreshBuffer is how long before expiry we should refresh the session.
	SessionRefreshBuffer = 5 * time.Minute

	// DefaultRetryMax is the default number of retries for transient errors.
	DefaultRetryMax = 3

	// DefaultRetryWaitMin is the minimum wait time between retries.
	DefaultRetryWaitMin = 2 * time.Second

	// DefaultRetryWaitMax is the maximum wait time between retries.
	DefaultRetryWaitMax = 10 * time.Second
)

// Client is a Pi-hole FTL API client.
type Client struct {
	baseURL    *url.URL
	httpClient *retryablehttp.Client
	password   string

	// Session management
	mu        sync.RWMutex
	sid       string
	sidExpiry time.Time
}

// Config holds the configuration for creating a new Client.
type Config struct {
	// URL is the base URL of the Pi-hole instance (e.g., "http://pi.hole").
	URL string

	// Password is the Pi-hole web interface password.
	Password string

	// TLSInsecureSkipVerify skips TLS certificate verification.
	TLSInsecureSkipVerify bool

	// Timeout is the HTTP timeout for API requests.
	Timeout time.Duration

	// RetryMax is the maximum number of retries for transient errors.
	RetryMax int

	// RetryWaitMin is the minimum wait time between retries.
	RetryWaitMin time.Duration

	// RetryWaitMax is the maximum wait time between retries.
	RetryWaitMax time.Duration
}

// New creates a new Pi-hole API client with automatic retry support.
func New(cfg Config) (*Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("Pi-hole URL is required")
	}

	baseURL, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid Pi-hole URL: %w", err)
	}

	// Ensure the URL has a path for the API
	if baseURL.Path == "" {
		baseURL.Path = "/api"
	} else if baseURL.Path[len(baseURL.Path)-1] != '/' {
		baseURL.Path += "/api"
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	retryMax := cfg.RetryMax
	if retryMax < 0 {
		// Negative value means explicitly disable retries
		retryMax = 0
	} else if retryMax == 0 {
		retryMax = DefaultRetryMax
	}

	retryWaitMin := cfg.RetryWaitMin
	if retryWaitMin == 0 {
		retryWaitMin = DefaultRetryWaitMin
	}

	retryWaitMax := cfg.RetryWaitMax
	if retryWaitMax == 0 {
		retryWaitMax = DefaultRetryWaitMax
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.TLSInsecureSkipVerify,
		},
	}

	// Create retryable HTTP client
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
	retryClient.RetryMax = retryMax
	retryClient.RetryWaitMin = retryWaitMin
	retryClient.RetryWaitMax = retryWaitMax
	retryClient.Logger = nil // Disable default noisy logging

	// Custom retry policy: retry on connection errors and 5xx
	retryClient.CheckRetry = retryablehttp.DefaultRetryPolicy

	return &Client{
		baseURL:    baseURL,
		password:   cfg.Password,
		httpClient: retryClient,
	}, nil
}

// AuthResponse represents the response from the authentication endpoint.
type AuthResponse struct {
	Session struct {
		Valid    bool   `json:"valid"`
		TOTP     bool   `json:"totp"`
		SID      string `json:"sid"`
		CSRF     string `json:"csrf"`
		Validity int    `json:"validity"` // seconds until expiry
		Message  string `json:"message"`
	} `json:"session"`
	Took float64 `json:"took"`
}

// ErrorResponse represents an error response from the API.
type ErrorResponse struct {
	Error struct {
		Key     string  `json:"key"`
		Message string  `json:"message"`
		Hint    *string `json:"hint"`
	} `json:"error"`
	Took float64 `json:"took"`
}

// Authenticate obtains a new session ID from Pi-hole.
func (c *Client) Authenticate(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.authenticateLocked(ctx)
}

func (c *Client) authenticateLocked(ctx context.Context) error {
	// First, check if authentication is required
	authURL := c.baseURL.JoinPath("auth")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create auth check request: %w", err)
	}

	retryReq, err := retryablehttp.FromRequest(req)
	if err != nil {
		return fmt.Errorf("failed to create retryable request: %w", err)
	}

	resp, err := c.httpClient.Do(retryReq)
	if err != nil {
		return fmt.Errorf("failed to check auth status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	// If session is already valid (no password set on Pi-hole), we're done
	if authResp.Session.Valid {
		c.sid = authResp.Session.SID
		c.sidExpiry = time.Now().Add(time.Duration(authResp.Session.Validity) * time.Second)
		return nil
	}

	// Need to authenticate with password
	if c.password == "" {
		return fmt.Errorf("authentication required but no password provided")
	}

	loginPayload := map[string]string{
		"password": c.password,
	}

	payloadBytes, err := json.Marshal(loginPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal login payload: %w", err)
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, authURL.String(), bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	retryReq, err = retryablehttp.FromRequest(req)
	if err != nil {
		return fmt.Errorf("failed to create retryable request: %w", err)
	}

	resp, err = c.httpClient.Do(retryReq)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read login response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			return fmt.Errorf("authentication failed: %s", errResp.Error.Message)
		}
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	if !authResp.Session.Valid {
		return fmt.Errorf("authentication failed: invalid session")
	}

	c.sid = authResp.Session.SID
	c.sidExpiry = time.Now().Add(time.Duration(authResp.Session.Validity) * time.Second)

	return nil
}

// ensureAuthenticated ensures we have a valid session, refreshing if needed.
func (c *Client) ensureAuthenticated(ctx context.Context) error {
	c.mu.RLock()
	valid := c.sid != "" && time.Now().Add(SessionRefreshBuffer).Before(c.sidExpiry)
	c.mu.RUnlock()

	if valid {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.sid != "" && time.Now().Add(SessionRefreshBuffer).Before(c.sidExpiry) {
		return nil
	}

	return c.authenticateLocked(ctx)
}

// Request makes an authenticated API request.
func (c *Client) Request(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	if err := c.ensureAuthenticated(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Parse the path to separate query string from path component
	var reqURL *url.URL
	if pathPart, queryPart, found := strings.Cut(path, "?"); found {
		reqURL = c.baseURL.JoinPath(pathPart)
		reqURL.RawQuery = queryPart
	} else {
		reqURL = c.baseURL.JoinPath(path)
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.mu.RLock()
	sid := c.sid
	c.mu.RUnlock()

	req.Header.Set("sid", sid)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	retryReq, err := retryablehttp.FromRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create retryable request: %w", err)
	}

	resp, err := c.httpClient.Do(retryReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
			hint := ""
			if errResp.Error.Hint != nil {
				hint = fmt.Sprintf(" (hint: %s)", *errResp.Error.Hint)
			}
			return nil, fmt.Errorf("API error [%s]: %s%s", errResp.Error.Key, errResp.Error.Message, hint)
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Get performs an authenticated GET request.
func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	return c.Request(ctx, http.MethodGet, path, nil)
}

// Post performs an authenticated POST request.
func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.Request(ctx, http.MethodPost, path, body)
}

// Put performs an authenticated PUT request.
func (c *Client) Put(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.Request(ctx, http.MethodPut, path, body)
}

// Delete performs an authenticated DELETE request.
func (c *Client) Delete(ctx context.Context, path string) ([]byte, error) {
	return c.Request(ctx, http.MethodDelete, path, nil)
}

// Patch performs an authenticated PATCH request.
func (c *Client) Patch(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.Request(ctx, http.MethodPatch, path, body)
}
