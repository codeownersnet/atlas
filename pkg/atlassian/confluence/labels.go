package confluence

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetLabels retrieves labels for content
func (c *Client) GetLabels(ctx context.Context, contentID string, prefix string, limit int) ([]Label, error) {
	path := fmt.Sprintf("%s/content/%s/label", c.getAPIPath(), contentID)

	params := make(map[string]string)
	if prefix != "" {
		params["prefix"] = prefix
	}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}

	path = buildURL(path, params)

	var response struct {
		Results []Label `json:"results"`
		Start   int     `json:"start"`
		Limit   int     `json:"limit"`
		Size    int     `json:"size"`
	}

	if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get labels for content %s: %w", contentID, err)
	}

	return response.Results, nil
}

// AddLabel adds a label to content
func (c *Client) AddLabel(ctx context.Context, contentID string, name string, prefix string) (*Label, error) {
	path := fmt.Sprintf("%s/content/%s/label", c.getAPIPath(), contentID)

	req := CreateLabelRequest{
		Name:   name,
		Prefix: prefix,
	}

	reqBody, err := json.Marshal([]CreateLabelRequest{req})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal label request: %w", err)
	}

	var labels []Label
	if err := c.doRequest(ctx, "POST", path, reqBody, &labels); err != nil {
		return nil, fmt.Errorf("failed to add label to content %s: %w", contentID, err)
	}

	if len(labels) > 0 {
		return &labels[0], nil
	}

	return &Label{Name: name, Prefix: prefix}, nil
}

// AddLabels adds multiple labels to content
func (c *Client) AddLabels(ctx context.Context, contentID string, labels []CreateLabelRequest) ([]Label, error) {
	path := fmt.Sprintf("%s/content/%s/label", c.getAPIPath(), contentID)

	reqBody, err := json.Marshal(labels)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal labels request: %w", err)
	}

	var result []Label
	if err := c.doRequest(ctx, "POST", path, reqBody, &result); err != nil {
		return nil, fmt.Errorf("failed to add labels to content %s: %w", contentID, err)
	}

	return result, nil
}

// RemoveLabel removes a label from content
func (c *Client) RemoveLabel(ctx context.Context, contentID string, labelName string) error {
	path := fmt.Sprintf("%s/content/%s/label/%s", c.getAPIPath(), contentID, labelName)

	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to remove label %s from content %s: %w", labelName, contentID, err)
	}

	return nil
}

// SearchByLabel searches for content by label
func (c *Client) SearchByLabel(ctx context.Context, labelName string, spaceKey string, limit int) ([]Content, error) {
	cql := fmt.Sprintf("label=\"%s\"", labelName)
	if spaceKey != "" {
		cql = fmt.Sprintf("%s and space=\"%s\"", cql, spaceKey)
	}

	result, err := c.SearchCQL(ctx, cql, &SearchOptions{
		Limit: limit,
	})
	if err != nil {
		return nil, err
	}

	return result.Results, nil
}
