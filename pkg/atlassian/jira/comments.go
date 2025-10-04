package jira

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetComments retrieves all comments for an issue
func (c *Client) GetComments(ctx context.Context, issueKey string) ([]Comment, error) {
	path := fmt.Sprintf("%s/issue/%s/comment", c.getAPIPath(), issueKey)

	var response Comments
	if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get comments for issue %s: %w", issueKey, err)
	}

	return response.Comments, nil
}

// GetComment retrieves a specific comment by ID
func (c *Client) GetComment(ctx context.Context, issueKey string, commentID string) (*Comment, error) {
	path := fmt.Sprintf("%s/issue/%s/comment/%s", c.getAPIPath(), issueKey, commentID)

	var comment Comment
	if err := c.doRequest(ctx, "GET", path, nil, &comment); err != nil {
		return nil, fmt.Errorf("failed to get comment %s for issue %s: %w", commentID, issueKey, err)
	}

	return &comment, nil
}

// AddComment adds a comment to an issue
func (c *Client) AddComment(ctx context.Context, issueKey string, body string, visibility *Visibility) (*Comment, error) {
	path := fmt.Sprintf("%s/issue/%s/comment", c.getAPIPath(), issueKey)

	request := CreateCommentRequest{
		Body:       body,
		Visibility: visibility,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal comment request: %w", err)
	}

	var comment Comment
	if err := c.doRequest(ctx, "POST", path, reqBody, &comment); err != nil {
		return nil, fmt.Errorf("failed to add comment to issue %s: %w", issueKey, err)
	}

	return &comment, nil
}

// UpdateComment updates an existing comment
func (c *Client) UpdateComment(ctx context.Context, issueKey string, commentID string, body string, visibility *Visibility) (*Comment, error) {
	path := fmt.Sprintf("%s/issue/%s/comment/%s", c.getAPIPath(), issueKey, commentID)

	request := CreateCommentRequest{
		Body:       body,
		Visibility: visibility,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal comment request: %w", err)
	}

	var comment Comment
	if err := c.doRequest(ctx, "PUT", path, reqBody, &comment); err != nil {
		return nil, fmt.Errorf("failed to update comment %s on issue %s: %w", commentID, issueKey, err)
	}

	return &comment, nil
}

// DeleteComment deletes a comment
func (c *Client) DeleteComment(ctx context.Context, issueKey string, commentID string) error {
	path := fmt.Sprintf("%s/issue/%s/comment/%s", c.getAPIPath(), issueKey, commentID)

	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete comment %s from issue %s: %w", commentID, issueKey, err)
	}

	return nil
}
