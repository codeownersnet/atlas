package jira

import (
	"context"
	"fmt"
	"strings"
)

// GetProjectsOptions contains options for listing projects
type GetProjectsOptions struct {
	Expand     []string // Resources to expand (e.g., "description", "lead", "issueTypes")
	Recent     int      // Number of recent projects to return
	Properties []string // Project properties to return
	StartAt    int      // Starting index for pagination (Cloud only)
	MaxResults int      // Maximum results per page (Cloud only)
}

// GetAllProjects retrieves all accessible projects
func (c *Client) GetAllProjects(ctx context.Context, opts *GetProjectsOptions) ([]Project, error) {
	path := c.getProjectSearchAPIPath()

	// Build query parameters
	params := make(map[string]string)
	if opts != nil {
		if len(opts.Expand) > 0 {
			params["expand"] = strings.Join(opts.Expand, ",")
		}
		if opts.Recent > 0 {
			params["recent"] = fmt.Sprintf("%d", opts.Recent)
		}
		if len(opts.Properties) > 0 {
			params["properties"] = strings.Join(opts.Properties, ",")
		}
		// Add pagination support
		if opts.StartAt > 0 {
			params["startAt"] = fmt.Sprintf("%d", opts.StartAt)
		}
		if opts.MaxResults > 0 {
			params["maxResults"] = fmt.Sprintf("%d", opts.MaxResults)
		}
	}

	path = buildURL(path, params)

	// Handle different response formats
	if c.IsCloud() {
		// Cloud v3 returns paginated response
		var response struct {
			Values     []Project `json:"values"`
			MaxResults int       `json:"maxResults"`
			StartAt    int       `json:"startAt"`
			Total      int       `json:"total"`
			IsLast     bool      `json:"isLast"`
		}
		if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
			return nil, fmt.Errorf("failed to get projects: %w", err)
		}
		return response.Values, nil
	}

	// Server v2 returns direct array
	var projects []Project
	if err := c.doRequest(ctx, "GET", path, nil, &projects); err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return projects, nil
}

// GetProject retrieves a project by key or ID
func (c *Client) GetProject(ctx context.Context, projectKey string, expand []string) (*Project, error) {
	path := fmt.Sprintf("%s/%s", c.getProjectAPIPath(), projectKey)

	// Build query parameters
	params := make(map[string]string)
	if len(expand) > 0 {
		params["expand"] = strings.Join(expand, ",")
	}

	path = buildURL(path, params)

	var project Project
	if err := c.doRequest(ctx, "GET", path, nil, &project); err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", projectKey, err)
	}

	return &project, nil
}

// GetProjectVersions retrieves all versions for a project
func (c *Client) GetProjectVersions(ctx context.Context, projectKey string) ([]Version, error) {
	path := fmt.Sprintf("%s/%s/versions", c.getProjectAPIPath(), projectKey)

	var versions []Version
	if err := c.doRequest(ctx, "GET", path, nil, &versions); err != nil {
		return nil, fmt.Errorf("failed to get versions for project %s: %w", projectKey, err)
	}

	return versions, nil
}

// GetProjectComponents retrieves all components for a project
func (c *Client) GetProjectComponents(ctx context.Context, projectKey string) ([]Component, error) {
	path := fmt.Sprintf("%s/%s/components", c.getProjectAPIPath(), projectKey)

	var components []Component
	if err := c.doRequest(ctx, "GET", path, nil, &components); err != nil {
		return nil, fmt.Errorf("failed to get components for project %s: %w", projectKey, err)
	}

	return components, nil
}

// GetProjectIssueTypes retrieves all issue types for a project
func (c *Client) GetProjectIssueTypes(ctx context.Context, projectKey string) ([]IssueType, error) {
	// Get project with issue types expanded
	project, err := c.GetProject(ctx, projectKey, []string{"issueTypes"})
	if err != nil {
		return nil, err
	}

	return project.IssueTypes, nil
}

// SearchProjects searches for projects using a query string
func (c *Client) SearchProjects(ctx context.Context, query string, maxResults int) ([]Project, error) {
	path := c.getProjectSearchAPIPath()

	params := make(map[string]string)
	if query != "" {
		params["query"] = query
	}
	if maxResults > 0 {
		params["maxResults"] = fmt.Sprintf("%d", maxResults)
	}

	path = buildURL(path, params)

	var response struct {
		Values []Project `json:"values"`
	}

	if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to search projects: %w", err)
	}

	return response.Values, nil
}
