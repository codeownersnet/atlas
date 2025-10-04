package jira

import (
	"context"
	"encoding/json"
	"fmt"
)

// CreateVersion creates a new version
func (c *Client) CreateVersion(ctx context.Context, req *CreateVersionRequest) (*Version, error) {
	path := fmt.Sprintf("%s/version", c.getAPIPath())

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal version request: %w", err)
	}

	var version Version
	if err := c.doRequest(ctx, "POST", path, reqBody, &version); err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	return &version, nil
}

// BatchCreateVersions creates multiple versions in a single request
func (c *Client) BatchCreateVersions(ctx context.Context, versions []*CreateVersionRequest) ([]Version, error) {
	var createdVersions []Version

	for _, versionReq := range versions {
		version, err := c.CreateVersion(ctx, versionReq)
		if err != nil {
			return createdVersions, fmt.Errorf("failed to create version %s: %w", versionReq.Name, err)
		}
		createdVersions = append(createdVersions, *version)
	}

	return createdVersions, nil
}

// GetVersion retrieves a version by ID
func (c *Client) GetVersion(ctx context.Context, versionID string) (*Version, error) {
	path := fmt.Sprintf("%s/version/%s", c.getAPIPath(), versionID)

	var version Version
	if err := c.doRequest(ctx, "GET", path, nil, &version); err != nil {
		return nil, fmt.Errorf("failed to get version %s: %w", versionID, err)
	}

	return &version, nil
}

// UpdateVersion updates an existing version
func (c *Client) UpdateVersion(ctx context.Context, versionID string, req *CreateVersionRequest) (*Version, error) {
	path := fmt.Sprintf("%s/version/%s", c.getAPIPath(), versionID)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal version request: %w", err)
	}

	var version Version
	if err := c.doRequest(ctx, "PUT", path, reqBody, &version); err != nil {
		return nil, fmt.Errorf("failed to update version %s: %w", versionID, err)
	}

	return &version, nil
}

// DeleteVersion deletes a version
func (c *Client) DeleteVersion(ctx context.Context, versionID string, moveFixIssuesTo, moveAffectedIssuesTo string) error {
	path := fmt.Sprintf("%s/version/%s", c.getAPIPath(), versionID)

	// Add query parameters for moving issues
	params := make(map[string]string)
	if moveFixIssuesTo != "" {
		params["moveFixIssuesTo"] = moveFixIssuesTo
	}
	if moveAffectedIssuesTo != "" {
		params["moveAffectedIssuesTo"] = moveAffectedIssuesTo
	}

	path = buildURL(path, params)

	if err := c.doRequest(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("failed to delete version %s: %w", versionID, err)
	}

	return nil
}

// ReleaseVersion marks a version as released
func (c *Client) ReleaseVersion(ctx context.Context, versionID string, releaseDate string) (*Version, error) {
	req := &CreateVersionRequest{
		Released:    true,
		ReleaseDate: releaseDate,
	}
	return c.UpdateVersion(ctx, versionID, req)
}

// ArchiveVersion marks a version as archived
func (c *Client) ArchiveVersion(ctx context.Context, versionID string) (*Version, error) {
	req := &CreateVersionRequest{
		Archived: true,
	}
	return c.UpdateVersion(ctx, versionID, req)
}
