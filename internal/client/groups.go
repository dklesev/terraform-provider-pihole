// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// GetGroups retrieves all groups or a specific group by name.
func (c *Client) GetGroups(ctx context.Context, name string) ([]Group, error) {
	path := "groups"
	if name != "" {
		path = fmt.Sprintf("groups/%s", url.PathEscape(name))
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result GroupsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse groups response: %w", err)
	}

	return result.Groups, nil
}

// GetGroup retrieves a specific group by name.
func (c *Client) GetGroup(ctx context.Context, name string) (*Group, error) {
	groups, err := c.GetGroups(ctx, name)
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return nil, nil // Not found
	}

	return &groups[0], nil
}

// CreateGroup creates a new group.
func (c *Client) CreateGroup(ctx context.Context, group *Group) (*Group, error) {
	payload := map[string]interface{}{
		"name":    group.Name,
		"enabled": group.Enabled,
	}
	if group.Description != "" {
		payload["comment"] = group.Description
	}

	resp, err := c.Post(ctx, "groups", payload)
	if err != nil {
		return nil, err
	}

	var result GroupsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse create group response: %w", err)
	}

	if len(result.Groups) == 0 {
		return nil, fmt.Errorf("no group returned in response")
	}

	return &result.Groups[0], nil
}

// UpdateGroup updates an existing group.
func (c *Client) UpdateGroup(ctx context.Context, name string, group *Group) (*Group, error) {
	payload := map[string]interface{}{
		"name":    group.Name,
		"enabled": group.Enabled,
		"comment": group.Description,
	}

	path := fmt.Sprintf("groups/%s", url.PathEscape(name))
	resp, err := c.Put(ctx, path, payload)
	if err != nil {
		return nil, err
	}

	var result GroupsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse update group response: %w", err)
	}

	// If groups array is empty (happens on rename), fetch the group by new name
	if len(result.Groups) == 0 {
		// Check if the update was successful via processed field
		// If rename succeeded, fetch the group by its new name
		updatedGroup, err := c.GetGroup(ctx, group.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch renamed group: %w", err)
		}
		if updatedGroup == nil {
			return nil, fmt.Errorf("no group returned in response and renamed group not found")
		}
		return updatedGroup, nil
	}

	return &result.Groups[0], nil
}

// DeleteGroup deletes a group by name.
func (c *Client) DeleteGroup(ctx context.Context, name string) error {
	path := fmt.Sprintf("groups/%s", url.PathEscape(name))
	_, err := c.Delete(ctx, path)
	return err
}
