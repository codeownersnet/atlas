package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// GetIssueOptions contains options for getting an issue
type GetIssueOptions struct {
	Fields     []string // Specific fields to retrieve, or "*all" for all fields
	Expand     []string // Resources to expand (e.g., "changelog", "renderedFields")
	Properties []string // Properties to retrieve
}

// GetIssue retrieves an issue by key or ID
func (c *Client) GetIssue(ctx context.Context, issueKey string, opts *GetIssueOptions) (*Issue, error) {
	path := fmt.Sprintf("%s/issue/%s", c.getAPIPath(), issueKey)

	// Build query parameters
	params := make(map[string]string)
	if opts != nil {
		if len(opts.Fields) > 0 {
			params["fields"] = strings.Join(opts.Fields, ",")
		}
		if len(opts.Expand) > 0 {
			params["expand"] = strings.Join(opts.Expand, ",")
		}
		if len(opts.Properties) > 0 {
			params["properties"] = strings.Join(opts.Properties, ",")
		}
	}

	path = buildURL(path, params)

	var issue Issue
	if err := c.doRequest(ctx, "GET", path, nil, &issue); err != nil {
		return nil, fmt.Errorf("failed to get issue %s: %w", issueKey, err)
	}

	return &issue, nil
}

// SearchOptions contains options for searching issues
type SearchOptions struct {
	Fields        []string // Specific fields to retrieve
	Expand        []string // Resources to expand
	StartAt       int      // Index of the first issue to return (0-based) - Server/DC only
	MaxResults    int      // Maximum number of issues to return (default 50, max 100 for Cloud, 1000 for Server)
	NextPageToken string   // Token for next page - Cloud v3 only
	ValidateQuery bool     // Whether to validate the JQL query
}

// SearchIssues searches for issues using JQL
func (c *Client) SearchIssues(ctx context.Context, jql string, opts *SearchOptions) (*SearchResult, error) {
	// Use the deployment-specific search endpoint
	// Cloud: /rest/api/3/search/jql (POST with JQL in body)
	// Server/DC: /rest/api/2/search (POST with JQL in body)
	path := c.getSearchAPIPath()

	// Build request body
	body := map[string]interface{}{
		"jql": jql,
	}

	if opts != nil {
		if len(opts.Fields) > 0 {
			// Handle "*all" differently for Cloud vs Server
			// Server v2: use "*all"
			// Cloud v3: must explicitly list fields (doesn't support "*all" or "*navigable")
			if len(opts.Fields) == 1 && opts.Fields[0] == "*all" {
				if c.IsCloud() {
					// For Cloud v3 API, explicitly list all standard fields
					// Custom fields (customfield_*) will also be returned
					body["fields"] = []string{
						"*all", // Try *all since we fixed the startAt issue
					}
				} else{
					// For Server, use "*all"
					body["fields"] = opts.Fields
				}
			} else {
				// For specific fields, add them normally
				body["fields"] = opts.Fields
			}
		}
		if len(opts.Expand) > 0 {
			body["expand"] = opts.Expand
		}

		// Handle pagination differently for Cloud vs Server
		// Cloud v3 uses token-based pagination (nextPageToken)
		// Server/DC uses offset-based pagination (startAt)
		if c.IsCloud() {
			// Cloud: use nextPageToken if provided (omit for first page)
			if opts.NextPageToken != "" {
				body["nextPageToken"] = opts.NextPageToken
			}
		} else {
			// Server: use startAt (always include it)
			body["startAt"] = opts.StartAt
		}

		if opts.MaxResults > 0 {
			body["maxResults"] = opts.MaxResults
		} else {
			body["maxResults"] = 50 // Default
		}
		if opts.ValidateQuery {
			body["validateQuery"] = true
		}
	} else {
		// No options provided - use defaults
		if !c.IsCloud() {
			body["startAt"] = 0 // Default for Server
		}
		body["maxResults"] = 50 // Default
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var result SearchResult
	if err := c.doRequest(ctx, "POST", path, reqBody, &result); err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	return &result, nil
}

// CreateIssue creates a new issue
func (c *Client) CreateIssue(ctx context.Context, fields map[string]interface{}) (*Issue, error) {
	path := fmt.Sprintf("%s/issue", c.getAPIPath())

	reqBody, err := json.Marshal(CreateIssueRequest{Fields: fields})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var issue Issue
	if err := c.doRequest(ctx, "POST", path, reqBody, &issue); err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return &issue, nil
}

// BatchCreateIssuesRequest represents a batch create request
type BatchCreateIssuesRequest struct {
	IssueUpdates []CreateIssueRequest `json:"issueUpdates"`
}

// BatchCreateIssuesResponse represents a batch create response
type BatchCreateIssuesResponse struct {
	Issues []Issue      `json:"issues,omitempty"`
	Errors []BatchError `json:"errors,omitempty"`
}

// BatchError represents an error in a batch operation
type BatchError struct {
	Status        int            `json:"status,omitempty"`
	ElementErrors *ErrorResponse `json:"elementErrors,omitempty"`
	FailedElement int            `json:"failedElementNumber,omitempty"`
}

// BatchCreateIssues creates multiple issues in a single request
func (c *Client) BatchCreateIssues(ctx context.Context, issuesFields []map[string]interface{}) (*BatchCreateIssuesResponse, error) {
	path := fmt.Sprintf("%s/issue/bulk", c.getAPIPath())

	issueUpdates := make([]CreateIssueRequest, len(issuesFields))
	for i, fields := range issuesFields {
		issueUpdates[i] = CreateIssueRequest{Fields: fields}
	}

	reqBody, err := json.Marshal(BatchCreateIssuesRequest{IssueUpdates: issueUpdates})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var result BatchCreateIssuesResponse
	if err := c.doRequest(ctx, "POST", path, reqBody, &result); err != nil {
		return nil, fmt.Errorf("failed to batch create issues: %w", err)
	}

	return &result, nil
}

// UpdateIssue updates an issue
func (c *Client) UpdateIssue(ctx context.Context, issueKey string, fields map[string]interface{}, update map[string]interface{}) error {
	path := fmt.Sprintf("%s/issue/%s", c.getAPIPath(), issueKey)

	reqBody, err := json.Marshal(UpdateIssueRequest{
		Fields: fields,
		Update: update,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	if err := c.doRequest(ctx, "PUT", path, reqBody, nil); err != nil {
		return fmt.Errorf("failed to update issue %s: %w", issueKey, err)
	}

	return nil
}

// DeleteIssue deletes an issue
func (c *Client) DeleteIssue(ctx context.Context, issueKey string, deleteSubtasks bool) error {
	path := fmt.Sprintf("%s/issue/%s", c.getAPIPath(), issueKey)

	if deleteSubtasks {
		path = fmt.Sprintf("%s?deleteSubtasks=true", path)
	}

	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete issue %s: %w", issueKey, err)
	}

	return nil
}

// AssignIssue assigns an issue to a user
func (c *Client) AssignIssue(ctx context.Context, issueKey string, accountID string) error {
	path := fmt.Sprintf("%s/issue/%s/assignee", c.getAPIPath(), issueKey)

	var reqBody []byte
	var err error

	if c.IsCloud() {
		reqBody, err = json.Marshal(map[string]string{"accountId": accountID})
	} else {
		reqBody, err = json.Marshal(map[string]string{"name": accountID})
	}

	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	if err := c.doRequest(ctx, "PUT", path, reqBody, nil); err != nil {
		return fmt.Errorf("failed to assign issue %s: %w", issueKey, err)
	}

	return nil
}

// GetChangelogs retrieves the changelog for an issue
func (c *Client) GetChangelogs(ctx context.Context, issueKey string) ([]Changelog, error) {
	opts := &GetIssueOptions{
		Expand: []string{"changelog"},
	}

	_, err := c.GetIssue(ctx, issueKey, opts)
	if err != nil {
		return nil, err
	}

	// Parse changelog from expand
	// Note: This is a simplified implementation
	// In a real scenario, we'd need to properly parse the expand data
	return []Changelog{}, nil
}

// BatchGetChangelogs retrieves changelogs for multiple issues (Cloud only)
func (c *Client) BatchGetChangelogs(ctx context.Context, issueKeys []string) (map[string][]Changelog, error) {
	if !c.IsCloud() {
		return nil, fmt.Errorf("batch changelog retrieval is only supported on Jira Cloud")
	}

	// This would typically use a batch endpoint
	// For now, we'll get them individually
	result := make(map[string][]Changelog)
	for _, key := range issueKeys {
		changelogs, err := c.GetChangelogs(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get changelogs for %s: %w", key, err)
		}
		result[key] = changelogs
	}

	return result, nil
}

// GetProjectIssues retrieves all issues for a project
func (c *Client) GetProjectIssues(ctx context.Context, projectKey string, opts *SearchOptions) (*SearchResult, error) {
	jql := fmt.Sprintf("project = %s ORDER BY created DESC", projectKey)
	return c.SearchIssues(ctx, jql, opts)
}

// LinkToEpic links an issue to an epic
func (c *Client) LinkToEpic(ctx context.Context, issueKey, epicKey string) error {
	// The epic link field varies between Cloud and Server
	// Cloud uses a special endpoint, Server uses a custom field

	if c.IsCloud() {
		// Cloud uses /rest/agile/1.0/epic/{epicKey}/issue
		path := fmt.Sprintf("%s/epic/%s/issue", c.getAgileAPIPath(), epicKey)
		reqBody, err := json.Marshal(map[string]interface{}{
			"issues": []string{issueKey},
		})
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		if err := c.doRequest(ctx, "POST", path, reqBody, nil); err != nil {
			return fmt.Errorf("failed to link issue %s to epic %s: %w", issueKey, epicKey, err)
		}
	} else {
		// Server/DC typically uses a custom field
		// The field name varies by installation, commonly "Epic Link" or "customfield_10014"
		// This would need to be configured or discovered
		return fmt.Errorf("epic linking on Server/DC requires custom field configuration")
	}

	return nil
}

// AddAttachment adds an attachment to an issue
func (c *Client) AddAttachment(ctx context.Context, issueKey string, filename string, content []byte) (*Attachment, error) {
	// This would require multipart/form-data handling
	// For now, this is a placeholder
	return nil, fmt.Errorf("attachment upload not yet implemented")
}

// DownloadAttachment downloads an attachment
func (c *Client) DownloadAttachment(ctx context.Context, attachmentURL string) ([]byte, error) {
	// Parse the attachment URL to get the path
	u, err := url.Parse(attachmentURL)
	if err != nil {
		return nil, fmt.Errorf("invalid attachment URL: %w", err)
	}

	resp, err := c.httpClient.Get(ctx, u.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to download attachment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to download attachment: HTTP %d", resp.StatusCode)
	}

	content := make([]byte, 0)
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			content = append(content, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	return content, nil
}
