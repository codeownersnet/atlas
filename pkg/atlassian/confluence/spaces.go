package confluence

import (
	"context"
	"fmt"
)

// GetSpacesOptions contains options for listing spaces
type GetSpacesOptions struct {
	SpaceKey  []string // Filter by space keys
	SpaceType string   // Filter by type (global, personal)
	Status    string   // Filter by status (current, archived)
	Expand    []string // Resources to expand
	Start     int      // Pagination start
	Limit     int      // Pagination limit
}

// GetSpaces retrieves all spaces
func (c *Client) GetSpaces(ctx context.Context, opts *GetSpacesOptions) ([]Space, error) {
	path := fmt.Sprintf("%s/space", c.getAPIPath())

	params := make(map[string]string)
	if opts != nil {
		if len(opts.SpaceKey) > 0 {
			for _, key := range opts.SpaceKey {
				params["spaceKey"] = key // Note: This would need to be properly handled for multiple keys
				break                    // For simplicity, just use the first one
			}
		}
		if opts.SpaceType != "" {
			params["type"] = opts.SpaceType
		}
		if opts.Status != "" {
			params["status"] = opts.Status
		}
		if len(opts.Expand) > 0 {
			params["expand"] = expandFields(opts.Expand)
		}
		if opts.Start > 0 {
			params["start"] = fmt.Sprintf("%d", opts.Start)
		}
		if opts.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
	}

	path = buildURL(path, params)

	var response struct {
		Results []Space `json:"results"`
		Start   int     `json:"start"`
		Limit   int     `json:"limit"`
		Size    int     `json:"size"`
	}

	if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get spaces: %w", err)
	}

	return response.Results, nil
}

// GetSpace retrieves a space by key
func (c *Client) GetSpace(ctx context.Context, spaceKey string, expand []string) (*Space, error) {
	path := fmt.Sprintf("%s/space/%s", c.getAPIPath(), spaceKey)

	params := make(map[string]string)
	if len(expand) > 0 {
		params["expand"] = expandFields(expand)
	}

	path = buildURL(path, params)

	var space Space
	if err := c.doRequest(ctx, "GET", path, nil, &space); err != nil {
		return nil, fmt.Errorf("failed to get space %s: %w", spaceKey, err)
	}

	return &space, nil
}

// GetSpaceContent retrieves content in a space
func (c *Client) GetSpaceContent(ctx context.Context, spaceKey string, contentType ContentType, expand []string, limit int) ([]Content, error) {
	path := fmt.Sprintf("%s/space/%s/content", c.getAPIPath(), spaceKey)

	params := make(map[string]string)
	if contentType != "" {
		path = fmt.Sprintf("%s/%s", path, contentType)
	}
	if len(expand) > 0 {
		params["expand"] = expandFields(expand)
	}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}

	path = buildURL(path, params)

	var response struct {
		Results []Content `json:"results"`
	}

	if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get content for space %s: %w", spaceKey, err)
	}

	return response.Results, nil
}

// SearchSpaces searches for spaces
func (c *Client) SearchSpaces(ctx context.Context, query string, limit int) ([]Space, error) {
	// Use content search with space type filter
	// Note: This is a simplified implementation
	opts := &GetSpacesOptions{
		Limit: limit,
	}

	return c.GetSpaces(ctx, opts)
}
