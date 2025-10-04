package jira

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetBoardsOptions contains options for getting boards
type GetBoardsOptions struct {
	ProjectKeyOrID string
	BoardType      string // scrum, kanban, simple
	Name           string
	StartAt        int
	MaxResults     int
}

// GetBoards retrieves all boards
func (c *Client) GetBoards(ctx context.Context, opts *GetBoardsOptions) ([]Board, error) {
	path := fmt.Sprintf("%s/board", c.getAgileAPIPath())

	params := make(map[string]string)
	if opts != nil {
		if opts.ProjectKeyOrID != "" {
			params["projectKeyOrId"] = opts.ProjectKeyOrID
		}
		if opts.BoardType != "" {
			params["type"] = opts.BoardType
		}
		if opts.Name != "" {
			params["name"] = opts.Name
		}
		if opts.StartAt > 0 {
			params["startAt"] = fmt.Sprintf("%d", opts.StartAt)
		}
		if opts.MaxResults > 0 {
			params["maxResults"] = fmt.Sprintf("%d", opts.MaxResults)
		}
	}

	path = buildURL(path, params)

	var response struct {
		MaxResults int     `json:"maxResults"`
		StartAt    int     `json:"startAt"`
		Total      int     `json:"total"`
		IsLast     bool    `json:"isLast"`
		Values     []Board `json:"values"`
	}

	if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get boards: %w", err)
	}

	return response.Values, nil
}

// GetBoard retrieves a specific board by ID
func (c *Client) GetBoard(ctx context.Context, boardID int) (*Board, error) {
	path := fmt.Sprintf("%s/board/%d", c.getAgileAPIPath(), boardID)

	var board Board
	if err := c.doRequest(ctx, "GET", path, nil, &board); err != nil {
		return nil, fmt.Errorf("failed to get board %d: %w", boardID, err)
	}

	return &board, nil
}

// GetBoardIssues retrieves issues for a board
func (c *Client) GetBoardIssues(ctx context.Context, boardID int, opts *SearchOptions) (*SearchResult, error) {
	path := fmt.Sprintf("%s/board/%d/issue", c.getAgileAPIPath(), boardID)

	params := make(map[string]string)
	if opts != nil {
		if opts.StartAt > 0 {
			params["startAt"] = fmt.Sprintf("%d", opts.StartAt)
		}
		if opts.MaxResults > 0 {
			params["maxResults"] = fmt.Sprintf("%d", opts.MaxResults)
		}
	}

	path = buildURL(path, params)

	var result SearchResult
	if err := c.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get issues for board %d: %w", boardID, err)
	}

	return &result, nil
}

// GetBoardSprints retrieves sprints for a board
func (c *Client) GetBoardSprints(ctx context.Context, boardID int, state string) ([]Sprint, error) {
	path := fmt.Sprintf("%s/board/%d/sprint", c.getAgileAPIPath(), boardID)

	params := make(map[string]string)
	if state != "" {
		params["state"] = state
	}

	path = buildURL(path, params)

	var response struct {
		MaxResults int      `json:"maxResults"`
		StartAt    int      `json:"startAt"`
		IsLast     bool     `json:"isLast"`
		Values     []Sprint `json:"values"`
	}

	if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get sprints for board %d: %w", boardID, err)
	}

	return response.Values, nil
}

// GetSprint retrieves a specific sprint by ID
func (c *Client) GetSprint(ctx context.Context, sprintID int) (*Sprint, error) {
	path := fmt.Sprintf("%s/sprint/%d", c.getAgileAPIPath(), sprintID)

	var sprint Sprint
	if err := c.doRequest(ctx, "GET", path, nil, &sprint); err != nil {
		return nil, fmt.Errorf("failed to get sprint %d: %w", sprintID, err)
	}

	return &sprint, nil
}

// GetSprintIssues retrieves issues for a sprint
func (c *Client) GetSprintIssues(ctx context.Context, sprintID int, opts *SearchOptions) (*SearchResult, error) {
	path := fmt.Sprintf("%s/sprint/%d/issue", c.getAgileAPIPath(), sprintID)

	params := make(map[string]string)
	if opts != nil {
		if opts.StartAt > 0 {
			params["startAt"] = fmt.Sprintf("%d", opts.StartAt)
		}
		if opts.MaxResults > 0 {
			params["maxResults"] = fmt.Sprintf("%d", opts.MaxResults)
		}
	}

	path = buildURL(path, params)

	var result SearchResult
	if err := c.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get issues for sprint %d: %w", sprintID, err)
	}

	return &result, nil
}

// CreateSprint creates a new sprint
func (c *Client) CreateSprint(ctx context.Context, req *CreateSprintRequest) (*Sprint, error) {
	path := fmt.Sprintf("%s/sprint", c.getAgileAPIPath())

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sprint request: %w", err)
	}

	var sprint Sprint
	if err := c.doRequest(ctx, "POST", path, reqBody, &sprint); err != nil {
		return nil, fmt.Errorf("failed to create sprint: %w", err)
	}

	return &sprint, nil
}

// UpdateSprint updates an existing sprint
func (c *Client) UpdateSprint(ctx context.Context, sprintID int, req *UpdateSprintRequest) (*Sprint, error) {
	path := fmt.Sprintf("%s/sprint/%d", c.getAgileAPIPath(), sprintID)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sprint request: %w", err)
	}

	var sprint Sprint
	if err := c.doRequest(ctx, "POST", path, reqBody, &sprint); err != nil {
		return nil, fmt.Errorf("failed to update sprint %d: %w", sprintID, err)
	}

	return &sprint, nil
}

// DeleteSprint deletes a sprint
func (c *Client) DeleteSprint(ctx context.Context, sprintID int) error {
	path := fmt.Sprintf("%s/sprint/%d", c.getAgileAPIPath(), sprintID)

	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete sprint %d: %w", sprintID, err)
	}

	return nil
}

// StartSprint starts a sprint
func (c *Client) StartSprint(ctx context.Context, sprintID int) (*Sprint, error) {
	req := &UpdateSprintRequest{
		State: "active",
	}
	return c.UpdateSprint(ctx, sprintID, req)
}

// CloseSprint closes a sprint
func (c *Client) CloseSprint(ctx context.Context, sprintID int) (*Sprint, error) {
	req := &UpdateSprintRequest{
		State: "closed",
	}
	return c.UpdateSprint(ctx, sprintID, req)
}

// MoveIssuesToSprint moves issues to a sprint
func (c *Client) MoveIssuesToSprint(ctx context.Context, sprintID int, issueKeys []string) error {
	path := fmt.Sprintf("%s/sprint/%d/issue", c.getAgileAPIPath(), sprintID)

	request := map[string]interface{}{
		"issues": issueKeys,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	if err := c.doRequest(ctx, "POST", path, reqBody, nil); err != nil {
		return fmt.Errorf("failed to move issues to sprint %d: %w", sprintID, err)
	}

	return nil
}

// GetBacklogIssues retrieves backlog issues for a board
func (c *Client) GetBacklogIssues(ctx context.Context, boardID int, opts *SearchOptions) (*SearchResult, error) {
	path := fmt.Sprintf("%s/board/%d/backlog", c.getAgileAPIPath(), boardID)

	params := make(map[string]string)
	if opts != nil {
		if opts.StartAt > 0 {
			params["startAt"] = fmt.Sprintf("%d", opts.StartAt)
		}
		if opts.MaxResults > 0 {
			params["maxResults"] = fmt.Sprintf("%d", opts.MaxResults)
		}
	}

	path = buildURL(path, params)

	var result SearchResult
	if err := c.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get backlog issues for board %d: %w", boardID, err)
	}

	return &result, nil
}
