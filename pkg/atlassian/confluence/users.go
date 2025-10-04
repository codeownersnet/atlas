package confluence

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

// SearchUsers searches for users using CQL
func (c *Client) SearchUsers(ctx context.Context, cql string, limit int) ([]User, error) {
	path := fmt.Sprintf("%s/search/user", c.getAPIPath())

	params := map[string]string{
		"cql": cql,
	}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}

	path = buildURL(path, params)

	var response struct {
		Results []User `json:"results"`
		Start   int    `json:"start"`
		Limit   int    `json:"limit"`
		Size    int    `json:"size"`
	}

	if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return response.Results, nil
}

// SearchUsersByName searches for users by name or email
func (c *Client) SearchUsersByName(ctx context.Context, query string, limit int) ([]User, error) {
	// Build CQL query
	cql := fmt.Sprintf("user.fullname~\"%s\" or user.email~\"%s\"", query, query)
	return c.SearchUsers(ctx, cql, limit)
}

// GetCurrentUser retrieves the currently authenticated user
func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	path := fmt.Sprintf("%s/user/current", c.getAPIPath())

	var user User
	if err := c.doRequest(ctx, "GET", path, nil, &user); err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	return &user, nil
}

// GetAnonymousUser retrieves the anonymous user information
func (c *Client) GetAnonymousUser(ctx context.Context) (*User, error) {
	path := fmt.Sprintf("%s/user/anonymous", c.getAPIPath())

	var user User
	if err := c.doRequest(ctx, "GET", path, nil, &user); err != nil {
		return nil, fmt.Errorf("failed to get anonymous user: %w", err)
	}

	return &user, nil
}

// GetUserGroups retrieves groups for a user
func (c *Client) GetUserGroups(ctx context.Context, accountIDOrUsername string) ([]string, error) {
	path := fmt.Sprintf("%s/user/memberof", c.getAPIPath())

	params := make(map[string]string)
	if c.IsCloud() {
		params["accountId"] = accountIDOrUsername
	} else {
		params["username"] = accountIDOrUsername
	}

	path = buildURL(path, params)

	var response struct {
		Results []struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"results"`
	}

	if err := c.doRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	groups := make([]string, 0, len(response.Results))
	for _, group := range response.Results {
		groups = append(groups, group.Name)
	}

	return groups, nil
}
