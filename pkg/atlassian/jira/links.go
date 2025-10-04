package jira

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetIssueLinkTypes retrieves all available issue link types
func (c *Client) GetIssueLinkTypes(ctx context.Context) ([]IssueLinkType, error) {
	path := fmt.Sprintf("%s/issueLinkType", c.getAPIPath())

	var response struct {
		IssueLinkTypes []IssueLinkType `json:"issueLinkTypes"`
	}

	if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get issue link types: %w", err)
	}

	return response.IssueLinkTypes, nil
}

// CreateIssueLink creates a link between two issues
func (c *Client) CreateIssueLink(ctx context.Context, linkType IssueLinkType, inwardIssue, outwardIssue string, comment *Comment) (*IssueLink, error) {
	path := fmt.Sprintf("%s/issueLink", c.getAPIPath())

	request := CreateIssueLinkRequest{
		Type: linkType,
		InwardIssue: LinkIssueRef{
			Key: inwardIssue,
		},
		OutwardIssue: LinkIssueRef{
			Key: outwardIssue,
		},
		Comment: comment,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal link request: %w", err)
	}

	// Issue link creation returns 201 with no body
	if err := c.doRequest(ctx, "POST", path, reqBody, nil); err != nil {
		return nil, fmt.Errorf("failed to create issue link: %w", err)
	}

	// Return a basic link structure
	return &IssueLink{
		Type: linkType,
		InwardIssue: &LinkedIssue{
			Key: inwardIssue,
		},
		OutwardIssue: &LinkedIssue{
			Key: outwardIssue,
		},
	}, nil
}

// CreateIssueLinkByName creates a link using the link type name
func (c *Client) CreateIssueLinkByName(ctx context.Context, linkTypeName, inwardIssue, outwardIssue string, comment *Comment) (*IssueLink, error) {
	// Get all link types
	linkTypes, err := c.GetIssueLinkTypes(ctx)
	if err != nil {
		return nil, err
	}

	// Find the link type by name
	var linkType *IssueLinkType
	for _, lt := range linkTypes {
		if lt.Name == linkTypeName {
			linkType = &lt
			break
		}
	}

	if linkType == nil {
		return nil, fmt.Errorf("link type '%s' not found", linkTypeName)
	}

	return c.CreateIssueLink(ctx, *linkType, inwardIssue, outwardIssue, comment)
}

// DeleteIssueLink deletes an issue link
func (c *Client) DeleteIssueLink(ctx context.Context, linkID string) error {
	path := fmt.Sprintf("%s/issueLink/%s", c.getAPIPath(), linkID)

	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete issue link %s: %w", linkID, err)
	}

	return nil
}

// GetIssueLink retrieves an issue link by ID
func (c *Client) GetIssueLink(ctx context.Context, linkID string) (*IssueLink, error) {
	path := fmt.Sprintf("%s/issueLink/%s", c.getAPIPath(), linkID)

	var link IssueLink
	if err := c.doRequest(ctx, "GET", path, nil, &link); err != nil {
		return nil, fmt.Errorf("failed to get issue link %s: %w", linkID, err)
	}

	return &link, nil
}

// CreateRemoteLink creates a remote link to an issue
func (c *Client) CreateRemoteLink(ctx context.Context, issueKey string, link *RemoteLink) (*RemoteLink, error) {
	path := fmt.Sprintf("%s/issue/%s/remotelink", c.getAPIPath(), issueKey)

	reqBody, err := json.Marshal(link)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal remote link request: %w", err)
	}

	var result RemoteLink
	if err := c.doRequest(ctx, "POST", path, reqBody, &result); err != nil {
		return nil, fmt.Errorf("failed to create remote link for issue %s: %w", issueKey, err)
	}

	return &result, nil
}

// GetRemoteLinks retrieves all remote links for an issue
func (c *Client) GetRemoteLinks(ctx context.Context, issueKey string) ([]RemoteLink, error) {
	path := fmt.Sprintf("%s/issue/%s/remotelink", c.getAPIPath(), issueKey)

	var links []RemoteLink
	if err := c.doRequest(ctx, "GET", path, nil, &links); err != nil {
		return nil, fmt.Errorf("failed to get remote links for issue %s: %w", issueKey, err)
	}

	return links, nil
}

// DeleteRemoteLink deletes a remote link
func (c *Client) DeleteRemoteLink(ctx context.Context, issueKey string, linkID string) error {
	path := fmt.Sprintf("%s/issue/%s/remotelink/%s", c.getAPIPath(), issueKey, linkID)

	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete remote link %s from issue %s: %w", linkID, issueKey, err)
	}

	return nil
}
