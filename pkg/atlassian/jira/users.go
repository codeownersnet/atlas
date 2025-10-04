package jira

import (
	"context"
	"fmt"
)

// GetUser retrieves a user by account ID (Cloud) or username (Server/DC)
func (c *Client) GetUser(ctx context.Context, accountIDOrUsername string) (*User, error) {
	path := fmt.Sprintf("%s/user", c.getAPIPath())

	params := make(map[string]string)
	if c.IsCloud() {
		params["accountId"] = accountIDOrUsername
	} else {
		params["username"] = accountIDOrUsername
	}

	path = buildURL(path, params)

	var user User
	if err := c.doRequest(ctx, "GET", path, nil, &user); err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", accountIDOrUsername, err)
	}

	return &user, nil
}

// SearchUsers searches for users
func (c *Client) SearchUsers(ctx context.Context, query string, maxResults int) ([]User, error) {
	var path string
	params := make(map[string]string)

	if c.IsCloud() {
		path = fmt.Sprintf("%s/user/search", c.getAPIPath())
		params["query"] = query
	} else {
		path = fmt.Sprintf("%s/user/search", c.getAPIPath())
		params["username"] = query
	}

	if maxResults > 0 {
		params["maxResults"] = fmt.Sprintf("%d", maxResults)
	}

	path = buildURL(path, params)

	var users []User
	if err := c.doRequest(ctx, "GET", path, nil, &users); err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return users, nil
}

// GetCurrentUser retrieves the currently authenticated user
func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	path := fmt.Sprintf("%s/myself", c.getAPIPath())

	var user User
	if err := c.doRequest(ctx, "GET", path, nil, &user); err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	return &user, nil
}

// FindAssignableUsers searches for users that can be assigned to issues
func (c *Client) FindAssignableUsers(ctx context.Context, project, issueKey, query string, maxResults int) ([]User, error) {
	path := fmt.Sprintf("%s/user/assignable/search", c.getAPIPath())

	params := make(map[string]string)
	if c.IsCloud() {
		params["query"] = query
	} else {
		params["username"] = query
	}

	if project != "" {
		params["project"] = project
	}
	if issueKey != "" {
		params["issueKey"] = issueKey
	}
	if maxResults > 0 {
		params["maxResults"] = fmt.Sprintf("%d", maxResults)
	}

	path = buildURL(path, params)

	var users []User
	if err := c.doRequest(ctx, "GET", path, nil, &users); err != nil {
		return nil, fmt.Errorf("failed to find assignable users: %w", err)
	}

	return users, nil
}

// GetUserGroups retrieves groups for a user
func (c *Client) GetUserGroups(ctx context.Context, accountIDOrUsername string) ([]string, error) {
	path := fmt.Sprintf("%s/user/groups", c.getAPIPath())

	params := make(map[string]string)
	if c.IsCloud() {
		params["accountId"] = accountIDOrUsername
	} else {
		params["username"] = accountIDOrUsername
	}

	path = buildURL(path, params)

	var groups []string
	if err := c.doRequest(ctx, "GET", path, nil, &groups); err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	return groups, nil
}
