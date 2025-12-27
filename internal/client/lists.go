// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// GetLists retrieves all lists or a specific list.
func (c *Client) GetLists(ctx context.Context, listType, address string) ([]List, error) {
	path := "lists"
	if address != "" {
		path = fmt.Sprintf("lists/%s", url.PathEscape(address))
	}

	// Add type query parameter if specified
	if listType != "" {
		path = fmt.Sprintf("%s?type=%s", path, listType)
	}

	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var result ListsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse lists response: %w", err)
	}

	return result.Lists, nil
}

// GetList retrieves a specific list by address and type.
func (c *Client) GetList(ctx context.Context, listType, address string) (*List, error) {
	lists, err := c.GetLists(ctx, listType, address)
	if err != nil {
		return nil, err
	}

	// Find exact match
	for _, l := range lists {
		if l.Address == address && l.Type == listType {
			return &l, nil
		}
	}

	return nil, nil // Not found
}

// CreateList creates a new list.
func (c *Client) CreateList(ctx context.Context, list *List) (*List, error) {
	if list.Type == "" {
		return nil, fmt.Errorf("list type is required")
	}

	payload := map[string]interface{}{
		"address": list.Address,
		"enabled": list.Enabled,
	}
	if list.Comment != "" {
		payload["comment"] = list.Comment
	}
	if len(list.Groups) > 0 {
		payload["groups"] = list.Groups
	}

	path := fmt.Sprintf("lists?type=%s", list.Type)
	resp, err := c.Post(ctx, path, payload)
	if err != nil {
		return nil, err
	}

	var result ListsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse create list response: %w", err)
	}

	if len(result.Lists) == 0 {
		return nil, fmt.Errorf("no list returned in response")
	}

	return &result.Lists[0], nil
}

// UpdateList updates an existing list.
func (c *Client) UpdateList(ctx context.Context, originalType, originalAddress string, list *List) (*List, error) {
	payload := map[string]interface{}{
		"address": list.Address,
		"enabled": list.Enabled,
		"comment": list.Comment,
		"groups":  list.Groups,
	}

	path := fmt.Sprintf("lists/%s?type=%s", url.PathEscape(originalAddress), originalType)
	resp, err := c.Put(ctx, path, payload)
	if err != nil {
		return nil, err
	}

	var result ListsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse update list response: %w", err)
	}

	// If lists array is empty (may happen on address/type change), fetch by new values
	if len(result.Lists) == 0 {
		updatedList, err := c.GetList(ctx, list.Type, list.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch updated list: %w", err)
		}
		if updatedList == nil {
			return nil, fmt.Errorf("no list returned in response and updated list not found")
		}
		return updatedList, nil
	}

	return &result.Lists[0], nil
}

// DeleteList deletes a list.
func (c *Client) DeleteList(ctx context.Context, listType, address string) error {
	path := fmt.Sprintf("lists/%s?type=%s", url.PathEscape(address), listType)
	_, err := c.Delete(ctx, path)
	return err
}
