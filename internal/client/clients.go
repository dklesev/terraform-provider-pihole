// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// GetClients retrieves all clients or a specific client.
func (c *Client) GetClients(ctx context.Context, client string) ([]PiholeClient, error) {
	path := "clients"
	if client != "" {
		path = fmt.Sprintf("clients/%s", url.PathEscape(client))
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result ClientsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse clients response: %w", err)
	}

	return result.Clients, nil
}

// GetClient retrieves a specific client.
func (c *Client) GetClient(ctx context.Context, client string) (*PiholeClient, error) {
	clients, err := c.GetClients(ctx, client)
	if err != nil {
		return nil, err
	}

	if len(clients) == 0 {
		return nil, nil // Not found
	}

	return &clients[0], nil
}

// CreateClient creates a new client.
func (c *Client) CreateClient(ctx context.Context, client *PiholeClient) (*PiholeClient, error) {
	payload := map[string]interface{}{
		"client": client.Client,
	}
	if client.Comment != "" {
		payload["comment"] = client.Comment
	}
	if len(client.Groups) > 0 {
		payload["groups"] = client.Groups
	}

	resp, err := c.Post(ctx, "clients", payload)
	if err != nil {
		return nil, err
	}

	var result ClientsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse create client response: %w", err)
	}

	if len(result.Clients) == 0 {
		return nil, fmt.Errorf("no client returned in response")
	}

	return &result.Clients[0], nil
}

// UpdateClient updates an existing client.
func (c *Client) UpdateClient(ctx context.Context, originalClient string, client *PiholeClient) (*PiholeClient, error) {
	payload := map[string]interface{}{
		"client":  client.Client,
		"comment": client.Comment,
		"groups":  client.Groups,
	}

	path := fmt.Sprintf("clients/%s", url.PathEscape(originalClient))
	resp, err := c.Put(ctx, path, payload)
	if err != nil {
		return nil, err
	}

	var result ClientsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse update client response: %w", err)
	}

	// If clients array is empty (may happen on client ID change), fetch by new ID
	if len(result.Clients) == 0 {
		updatedClient, err := c.GetClient(ctx, client.Client)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch updated client: %w", err)
		}
		if updatedClient == nil {
			return nil, fmt.Errorf("no client returned in response and updated client not found")
		}
		return updatedClient, nil
	}

	return &result.Clients[0], nil
}

// DeleteClient deletes a client.
func (c *Client) DeleteClient(ctx context.Context, client string) error {
	path := fmt.Sprintf("clients/%s", url.PathEscape(client))
	_, err := c.Delete(ctx, path)
	return err
}
