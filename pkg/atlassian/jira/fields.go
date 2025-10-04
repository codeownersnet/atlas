package jira

import (
	"context"
	"fmt"
	"strings"
)

// GetAllFields retrieves all fields (standard and custom)
func (c *Client) GetAllFields(ctx context.Context) ([]Field, error) {
	path := fmt.Sprintf("%s/field", c.getAPIPath())

	var fields []Field
	if err := c.doRequest(ctx, "GET", path, nil, &fields); err != nil {
		return nil, fmt.Errorf("failed to get fields: %w", err)
	}

	return fields, nil
}

// SearchFields searches for fields by name or key (fuzzy search)
func (c *Client) SearchFields(ctx context.Context, query string) ([]Field, error) {
	allFields, err := c.GetAllFields(ctx)
	if err != nil {
		return nil, err
	}

	if query == "" {
		return allFields, nil
	}

	query = strings.ToLower(query)
	var matches []Field

	for _, field := range allFields {
		// Check if query matches name, key, or ID
		if strings.Contains(strings.ToLower(field.Name), query) ||
			strings.Contains(strings.ToLower(field.Key), query) ||
			strings.Contains(strings.ToLower(field.ID), query) {
			matches = append(matches, field)
		}
	}

	return matches, nil
}

// GetFieldByName retrieves a field by its exact name
func (c *Client) GetFieldByName(ctx context.Context, name string) (*Field, error) {
	allFields, err := c.GetAllFields(ctx)
	if err != nil {
		return nil, err
	}

	for _, field := range allFields {
		if field.Name == name {
			return &field, nil
		}
	}

	return nil, fmt.Errorf("field '%s' not found", name)
}

// GetFieldByKey retrieves a field by its key or ID
func (c *Client) GetFieldByKey(ctx context.Context, key string) (*Field, error) {
	allFields, err := c.GetAllFields(ctx)
	if err != nil {
		return nil, err
	}

	for _, field := range allFields {
		if field.Key == key || field.ID == key {
			return &field, nil
		}
	}

	return nil, fmt.Errorf("field with key '%s' not found", key)
}

// GetCustomFields retrieves only custom fields
func (c *Client) GetCustomFields(ctx context.Context) ([]Field, error) {
	allFields, err := c.GetAllFields(ctx)
	if err != nil {
		return nil, err
	}

	var customFields []Field
	for _, field := range allFields {
		if field.Custom {
			customFields = append(customFields, field)
		}
	}

	return customFields, nil
}

// GetEssentialFields returns a list of essential field IDs
func (c *Client) GetEssentialFields() []string {
	return []string{
		"summary",
		"status",
		"issuetype",
		"project",
		"created",
		"updated",
		"priority",
		"assignee",
		"reporter",
		"description",
	}
}

// GetEpicLinkField attempts to find the Epic Link field
func (c *Client) GetEpicLinkField(ctx context.Context) (*Field, error) {
	allFields, err := c.GetAllFields(ctx)
	if err != nil {
		return nil, err
	}

	// Common names for Epic Link field
	epicLinkNames := []string{
		"Epic Link",
		"Epic link",
		"epic link",
		"Parent Link",
	}

	for _, field := range allFields {
		for _, name := range epicLinkNames {
			if field.Name == name {
				return &field, nil
			}
		}
		// Also check if it's a custom field with "epic" in the name
		if field.Custom && strings.Contains(strings.ToLower(field.Name), "epic") {
			return &field, nil
		}
	}

	return nil, fmt.Errorf("epic link field not found")
}

// GetStoryPointsField attempts to find the Story Points field
func (c *Client) GetStoryPointsField(ctx context.Context) (*Field, error) {
	allFields, err := c.GetAllFields(ctx)
	if err != nil {
		return nil, err
	}

	// Common names for Story Points field
	storyPointsNames := []string{
		"Story Points",
		"Story points",
		"story points",
		"Estimate",
	}

	for _, field := range allFields {
		for _, name := range storyPointsNames {
			if field.Name == name {
				return &field, nil
			}
		}
	}

	return nil, fmt.Errorf("story points field not found")
}

// ParseFieldList parses a field list string and returns the appropriate fields
// "*all" returns all fields, otherwise returns the specified fields
func (c *Client) ParseFieldList(ctx context.Context, fieldList string) ([]string, error) {
	if fieldList == "" {
		return c.GetEssentialFields(), nil
	}

	if fieldList == "*all" {
		return []string{"*all"}, nil
	}

	// Split comma-separated list
	fields := strings.Split(fieldList, ",")
	for i, f := range fields {
		fields[i] = strings.TrimSpace(f)
	}

	return fields, nil
}
