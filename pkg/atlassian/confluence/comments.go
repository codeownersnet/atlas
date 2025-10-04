package confluence

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetComments retrieves comments for a page
func (c *Client) GetComments(ctx context.Context, pageID string, expand []string, limit int) ([]Comment, error) {
	path := fmt.Sprintf("%s/content/%s/child/comment", c.getAPIPath(), pageID)

	params := make(map[string]string)
	if len(expand) > 0 {
		params["expand"] = expandFields(expand)
	}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}

	path = buildURL(path, params)

	var response struct {
		Results []Comment `json:"results"`
		Start   int       `json:"start"`
		Limit   int       `json:"limit"`
		Size    int       `json:"size"`
	}

	if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get comments for page %s: %w", pageID, err)
	}

	return response.Results, nil
}

// GetComment retrieves a specific comment by ID
func (c *Client) GetComment(ctx context.Context, commentID string, expand []string) (*Comment, error) {
	path := fmt.Sprintf("%s/content/%s", c.getAPIPath(), commentID)

	params := make(map[string]string)
	if len(expand) > 0 {
		params["expand"] = expandFields(expand)
	}

	path = buildURL(path, params)

	var comment Comment
	if err := c.doRequest(ctx, "GET", path, nil, &comment); err != nil {
		return nil, fmt.Errorf("failed to get comment %s: %w", commentID, err)
	}

	return &comment, nil
}

// AddComment adds a comment to a page
func (c *Client) AddComment(ctx context.Context, pageID string, body string) (*Comment, error) {
	path := fmt.Sprintf("%s/content", c.getAPIPath())

	req := CreateCommentRequest{
		Type: "comment",
		Container: &ContentRef{
			ID:   pageID,
			Type: "page",
		},
		Body: &Body{
			Storage: &BodyContent{
				Value:          body,
				Representation: FormatStorage,
			},
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal comment request: %w", err)
	}

	var comment Comment
	if err := c.doRequest(ctx, "POST", path, reqBody, &comment); err != nil {
		return nil, fmt.Errorf("failed to add comment to page %s: %w", pageID, err)
	}

	return &comment, nil
}

// UpdateComment updates an existing comment
func (c *Client) UpdateComment(ctx context.Context, commentID string, body string, version int) (*Comment, error) {
	path := fmt.Sprintf("%s/content/%s", c.getAPIPath(), commentID)

	req := UpdateContentRequest{
		Version: &Version{Number: version},
		Type:    "comment",
		Body: &Body{
			Storage: &BodyContent{
				Value:          body,
				Representation: FormatStorage,
			},
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal comment request: %w", err)
	}

	var comment Comment
	if err := c.doRequest(ctx, "PUT", path, reqBody, &comment); err != nil {
		return nil, fmt.Errorf("failed to update comment %s: %w", commentID, err)
	}

	return &comment, nil
}

// DeleteComment deletes a comment
func (c *Client) DeleteComment(ctx context.Context, commentID string) error {
	path := fmt.Sprintf("%s/content/%s", c.getAPIPath(), commentID)

	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete comment %s: %w", commentID, err)
	}

	return nil
}
