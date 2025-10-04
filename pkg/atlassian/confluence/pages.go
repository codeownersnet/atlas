package confluence

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// GetContentOptions contains options for getting content
type GetContentOptions struct {
	Expand  []string      // Resources to expand (e.g., "body.storage", "version", "space")
	Status  ContentStatus // Filter by status
	Version int           // Specific version number
}

// GetContent retrieves content by ID
func (c *Client) GetContent(ctx context.Context, contentID string, opts *GetContentOptions) (*Content, error) {
	path := fmt.Sprintf("%s/content/%s", c.getAPIPath(), contentID)

	params := make(map[string]string)
	if opts != nil {
		if len(opts.Expand) > 0 {
			params["expand"] = expandFields(opts.Expand)
		}
		if opts.Status != "" {
			params["status"] = string(opts.Status)
		}
		if opts.Version > 0 {
			params["version"] = fmt.Sprintf("%d", opts.Version)
		}
	}

	path = buildURL(path, params)

	var content Content
	if err := c.doRequest(ctx, "GET", path, nil, &content); err != nil {
		return nil, fmt.Errorf("failed to get content %s: %w", contentID, err)
	}

	return &content, nil
}

// GetPage retrieves a page by ID
func (c *Client) GetPage(ctx context.Context, pageID string, expand []string) (*Content, error) {
	return c.GetContent(ctx, pageID, &GetContentOptions{
		Expand: expand,
		Status: ContentStatusCurrent,
	})
}

// GetPageByTitle retrieves a page by title and space key
func (c *Client) GetPageByTitle(ctx context.Context, spaceKey, title string, expand []string) (*Content, error) {
	// Search for the page
	cql := fmt.Sprintf("type=page and space=%s and title=\"%s\"", spaceKey, title)
	results, err := c.SearchCQL(ctx, cql, &SearchOptions{
		Expand: expand,
		Limit:  1,
	})
	if err != nil {
		return nil, err
	}

	if len(results.Results) == 0 {
		return nil, fmt.Errorf("page not found: %s in space %s", title, spaceKey)
	}

	return &results.Results[0], nil
}

// CreateContent creates new content
func (c *Client) CreateContent(ctx context.Context, req *CreateContentRequest) (*Content, error) {
	path := fmt.Sprintf("%s/content", c.getAPIPath())

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var content Content
	if err := c.doRequest(ctx, "POST", path, reqBody, &content); err != nil {
		return nil, fmt.Errorf("failed to create content: %w", err)
	}

	return &content, nil
}

// CreatePage creates a new page
func (c *Client) CreatePage(ctx context.Context, spaceKey, title, body string, parentID string) (*Content, error) {
	req := &CreateContentRequest{
		Type:  ContentTypePage,
		Title: title,
		Space: &SpaceRef{Key: spaceKey},
		Body: &Body{
			Storage: &BodyContent{
				Value:          body,
				Representation: FormatStorage,
			},
		},
		Status: ContentStatusCurrent,
	}

	if parentID != "" {
		req.Ancestors = []ContentRef{{ID: parentID}}
	}

	return c.CreateContent(ctx, req)
}

// UpdateContent updates existing content
func (c *Client) UpdateContent(ctx context.Context, contentID string, req *UpdateContentRequest) (*Content, error) {
	path := fmt.Sprintf("%s/content/%s", c.getAPIPath(), contentID)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var content Content
	if err := c.doRequest(ctx, "PUT", path, reqBody, &content); err != nil {
		return nil, fmt.Errorf("failed to update content %s: %w", contentID, err)
	}

	return &content, nil
}

// UpdatePage updates an existing page
func (c *Client) UpdatePage(ctx context.Context, pageID string, title, body string, version int) (*Content, error) {
	req := &UpdateContentRequest{
		Version: &Version{Number: version},
		Title:   title,
		Type:    ContentTypePage,
		Body: &Body{
			Storage: &BodyContent{
				Value:          body,
				Representation: FormatStorage,
			},
		},
	}

	return c.UpdateContent(ctx, pageID, req)
}

// DeleteContent deletes content
func (c *Client) DeleteContent(ctx context.Context, contentID string) error {
	path := fmt.Sprintf("%s/content/%s", c.getAPIPath(), contentID)

	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete content %s: %w", contentID, err)
	}

	return nil
}

// DeletePage deletes a page
func (c *Client) DeletePage(ctx context.Context, pageID string) error {
	return c.DeleteContent(ctx, pageID)
}

// SearchOptions contains options for searching
type SearchOptions struct {
	Expand []string
	Start  int
	Limit  int
}

// SearchCQL searches content using CQL (Confluence Query Language)
func (c *Client) SearchCQL(ctx context.Context, cql string, opts *SearchOptions) (*SearchResult, error) {
	path := fmt.Sprintf("%s/content/search", c.getAPIPath())

	params := map[string]string{
		"cql": cql,
	}

	if opts != nil {
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

	var result SearchResult
	if err := c.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to search with CQL: %w", err)
	}

	return &result, nil
}

// Search searches content using text or CQL
func (c *Client) Search(ctx context.Context, query string, opts *SearchOptions) (*SearchResult, error) {
	// Auto-detect if it's CQL or simple text search
	if isCQL(query) {
		return c.SearchCQL(ctx, query, opts)
	}

	// Convert simple text to CQL
	cql := fmt.Sprintf("text ~ \"%s\"", query)
	return c.SearchCQL(ctx, cql, opts)
}

// isCQL checks if a query string is CQL
func isCQL(query string) bool {
	// Simple heuristic: if it contains '=' or known CQL operators, it's likely CQL
	cqlIndicators := []string{"=", "type=", "space=", "title~", "text~", "AND", "OR", "NOT"}
	for _, indicator := range cqlIndicators {
		if contains(query, indicator) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetPageChildren retrieves child pages of a page
func (c *Client) GetPageChildren(ctx context.Context, pageID string, expand []string, limit int) ([]Content, error) {
	path := fmt.Sprintf("%s/content/%s/child/page", c.getAPIPath(), pageID)

	params := make(map[string]string)
	if len(expand) > 0 {
		params["expand"] = expandFields(expand)
	}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}

	path = buildURL(path, params)

	var response ContentArray
	if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get children for page %s: %w", pageID, err)
	}

	return response.Results, nil
}

// GetPageAncestors retrieves ancestors of a page
func (c *Client) GetPageAncestors(ctx context.Context, pageID string) ([]Content, error) {
	page, err := c.GetPage(ctx, pageID, []string{"ancestors"})
	if err != nil {
		return nil, err
	}

	return page.Ancestors, nil
}

// GetPageHistory retrieves the history of a page
func (c *Client) GetPageHistory(ctx context.Context, pageID string) (*History, error) {
	page, err := c.GetPage(ctx, pageID, []string{"history"})
	if err != nil {
		return nil, err
	}

	return page.History, nil
}

// ConvertMarkdownToStorage converts Markdown to Confluence storage format
// This is a stub - actual implementation would use a proper converter
func (c *Client) ConvertMarkdownToStorage(ctx context.Context, markdown string) (string, error) {
	// TODO: Implement proper Markdown to Confluence storage format conversion
	// For now, return a basic HTML-like structure
	return fmt.Sprintf("<p>%s</p>", url.QueryEscape(markdown)), nil
}

// ConvertWikiToStorage converts Wiki markup to Confluence storage format
// This is a stub - actual implementation would use the Confluence API
func (c *Client) ConvertWikiToStorage(ctx context.Context, wiki string) (string, error) {
	// TODO: Implement Wiki to storage format conversion
	// This would typically use /rest/api/contentbody/convert endpoint
	return wiki, nil
}
