package jira

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetWorklogs retrieves all worklogs for an issue
func (c *Client) GetWorklogs(ctx context.Context, issueKey string) ([]Worklog, error) {
	path := fmt.Sprintf("%s/issue/%s/worklog", c.getAPIPath(), issueKey)

	var response Worklogs
	if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get worklogs for issue %s: %w", issueKey, err)
	}

	return response.Worklogs, nil
}

// GetWorklog retrieves a specific worklog by ID
func (c *Client) GetWorklog(ctx context.Context, issueKey string, worklogID string) (*Worklog, error) {
	path := fmt.Sprintf("%s/issue/%s/worklog/%s", c.getAPIPath(), issueKey, worklogID)

	var worklog Worklog
	if err := c.doRequest(ctx, "GET", path, nil, &worklog); err != nil {
		return nil, fmt.Errorf("failed to get worklog %s for issue %s: %w", worklogID, issueKey, err)
	}

	return &worklog, nil
}

// AddWorklog adds a worklog entry to an issue
func (c *Client) AddWorklog(ctx context.Context, issueKey string, req *CreateWorklogRequest) (*Worklog, error) {
	path := fmt.Sprintf("%s/issue/%s/worklog", c.getAPIPath(), issueKey)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal worklog request: %w", err)
	}

	var worklog Worklog
	if err := c.doRequest(ctx, "POST", path, reqBody, &worklog); err != nil {
		return nil, fmt.Errorf("failed to add worklog to issue %s: %w", issueKey, err)
	}

	return &worklog, nil
}

// UpdateWorklog updates an existing worklog
func (c *Client) UpdateWorklog(ctx context.Context, issueKey string, worklogID string, req *CreateWorklogRequest) (*Worklog, error) {
	path := fmt.Sprintf("%s/issue/%s/worklog/%s", c.getAPIPath(), issueKey, worklogID)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal worklog request: %w", err)
	}

	var worklog Worklog
	if err := c.doRequest(ctx, "PUT", path, reqBody, &worklog); err != nil {
		return nil, fmt.Errorf("failed to update worklog %s on issue %s: %w", worklogID, issueKey, err)
	}

	return &worklog, nil
}

// DeleteWorklog deletes a worklog
func (c *Client) DeleteWorklog(ctx context.Context, issueKey string, worklogID string) error {
	path := fmt.Sprintf("%s/issue/%s/worklog/%s", c.getAPIPath(), issueKey, worklogID)

	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete worklog %s from issue %s: %w", worklogID, issueKey, err)
	}

	return nil
}
