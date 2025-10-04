package jira

import (
	"context"
	"encoding/json"
	"fmt"
)

// TransitionsResponse represents the response from getting transitions
type TransitionsResponse struct {
	Expand      string       `json:"expand,omitempty"`
	Transitions []Transition `json:"transitions"`
}

// GetTransitions retrieves available transitions for an issue
func (c *Client) GetTransitions(ctx context.Context, issueKey string) ([]Transition, error) {
	path := fmt.Sprintf("%s/issue/%s/transitions", c.getAPIPath(), issueKey)

	var response TransitionsResponse
	if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get transitions for issue %s: %w", issueKey, err)
	}

	return response.Transitions, nil
}

// TransitionIssue transitions an issue to a new status
func (c *Client) TransitionIssue(ctx context.Context, issueKey string, transitionID string, fields map[string]interface{}) error {
	path := fmt.Sprintf("%s/issue/%s/transitions", c.getAPIPath(), issueKey)

	request := TransitionRequest{
		Transition: Transition{
			ID: transitionID,
		},
		Fields: fields,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal transition request: %w", err)
	}

	if err := c.doRequest(ctx, "POST", path, reqBody, nil); err != nil {
		return fmt.Errorf("failed to transition issue %s: %w", issueKey, err)
	}

	return nil
}

// TransitionIssueByName transitions an issue using the transition name instead of ID
func (c *Client) TransitionIssueByName(ctx context.Context, issueKey string, transitionName string, fields map[string]interface{}) error {
	// Get available transitions
	transitions, err := c.GetTransitions(ctx, issueKey)
	if err != nil {
		return err
	}

	// Find the transition by name
	var transitionID string
	for _, t := range transitions {
		if t.Name == transitionName {
			transitionID = t.ID
			break
		}
	}

	if transitionID == "" {
		return fmt.Errorf("transition '%s' not found for issue %s", transitionName, issueKey)
	}

	return c.TransitionIssue(ctx, issueKey, transitionID, fields)
}

// GetTransitionByName retrieves a specific transition by name
func (c *Client) GetTransitionByName(ctx context.Context, issueKey string, transitionName string) (*Transition, error) {
	transitions, err := c.GetTransitions(ctx, issueKey)
	if err != nil {
		return nil, err
	}

	for _, t := range transitions {
		if t.Name == transitionName {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("transition '%s' not found for issue %s", transitionName, issueKey)
}
