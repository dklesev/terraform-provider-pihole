// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetDNSBlocking gets the current DNS blocking status.
func (c *Client) GetDNSBlocking(ctx context.Context) (*DNSBlocking, error) {
	resp, err := c.Get(ctx, "dns/blocking")
	if err != nil {
		return nil, err
	}

	var result DNSBlockingResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse dns blocking response: %w", err)
	}

	return &DNSBlocking{
		Blocking: result.Blocking,
		Timer:    result.Timer,
	}, nil
}

// SetDNSBlocking sets the DNS blocking status.
func (c *Client) SetDNSBlocking(ctx context.Context, enabled bool, timer *float64) (*DNSBlocking, error) {
	payload := DNSBlockingRequest{
		Blocking: enabled,
		Timer:    timer,
	}

	resp, err := c.Post(ctx, "dns/blocking", payload)
	if err != nil {
		return nil, err
	}

	var result DNSBlockingResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse dns blocking response: %w", err)
	}

	return &DNSBlocking{
		Blocking: result.Blocking,
		Timer:    result.Timer,
	}, nil
}
