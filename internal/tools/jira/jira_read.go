package jira

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/codeownersnet/atlas/internal/mcp"
	"github.com/codeownersnet/atlas/pkg/atlassian/jira"
)

// JiraGetIssueTool creates the jira_get_issue tool
func JiraGetIssueTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_get_issue",
		"Get detailed information about a Jira issue by key or ID. Supports field filtering ('essential', '*all', or comma-separated field names) and relationship expansion.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"issue_key": mcp.NewStringProperty("Issue key (e.g., 'PROJ-123') or ID"),
				"fields": mcp.NewStringProperty("Fields to retrieve: 'essential' (default), '*all', or comma-separated field names (e.g., 'summary,status,assignee')").
					WithDefault("essential"),
				"expand": mcp.NewStringProperty("Resources to expand (e.g., 'changelog,renderedFields'). Comma-separated."),
			},
			"issue_key",
		),
		jiraGetIssueHandler,
		"jira", "read",
	)
}

func jiraGetIssueHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	issueKey, ok := args["issue_key"].(string)
	if !ok || issueKey == "" {
		return nil, fmt.Errorf("issue_key is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	opts := &jira.GetIssueOptions{}

	// Handle fields parameter
	if fields, ok := args["fields"].(string); ok && fields != "" {
		if fields == "*all" {
			opts.Fields = []string{"*all"}
		} else if fields != "essential" {
			opts.Fields = strings.Split(fields, ",")
		}
	}

	// Handle expand parameter
	if expand, ok := args["expand"].(string); ok && expand != "" {
		opts.Expand = strings.Split(expand, ",")
	}

	issue, err := client.GetIssue(ctx, issueKey, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	return mcp.NewJSONResult(issue)
}

// JiraSearchTool creates the jira_search tool
func JiraSearchTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_search",
		"Search for Jira issues using JQL (Jira Query Language). Supports pagination and field filtering.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"jql": mcp.NewStringProperty("JQL query string (e.g., 'project = PROJ AND status = Open')"),
				"fields": mcp.NewStringProperty("Fields to retrieve: 'essential' (default), '*all', or comma-separated field names").
					WithDefault("essential"),
				"start_at": mcp.NewIntegerProperty("Starting index for pagination (0-based)").
					WithDefault(0),
				"max_results": mcp.NewIntegerProperty("Maximum number of results to return (default 50)").
					WithDefault(50),
			},
			"jql",
		),
		jiraSearchHandler,
		"jira", "read",
	)
}

func jiraSearchHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	jql, ok := args["jql"].(string)
	if !ok || jql == "" {
		return nil, fmt.Errorf("jql is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	opts := &jira.SearchOptions{
		StartAt:    getIntArg(args, "start_at", 0),
		MaxResults: getIntArg(args, "max_results", 50),
	}

	// Handle fields parameter
	if fields, ok := args["fields"].(string); ok && fields != "" {
		if fields == "*all" {
			opts.Fields = []string{"*all"}
		} else if fields == "essential" {
			// Essential fields for search results
			opts.Fields = []string{
				"summary", "status", "assignee", "reporter", "priority",
				"issuetype", "project", "created", "updated", "key",
			}
		} else {
			opts.Fields = strings.Split(fields, ",")
		}
	} else {
		// Default to essential fields if not specified
		opts.Fields = []string{
			"summary", "status", "assignee", "reporter", "priority",
			"issuetype", "project", "created", "updated", "key",
		}
	}

	result, err := client.SearchIssues(ctx, jql, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	return mcp.NewJSONResult(result)
}

// JiraSearchFieldsTool creates the jira_search_fields tool
func JiraSearchFieldsTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_search_fields",
		"Search and discover Jira fields including custom fields. Useful for finding field IDs and names.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"query": mcp.NewStringProperty("Search query to filter fields by name or ID (fuzzy matching)"),
			},
		),
		jiraSearchFieldsHandler,
		"jira", "read",
	)
}

func jiraSearchFieldsHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	fields, err := client.GetAllFields(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get fields: %w", err)
	}

	// Filter fields if query is provided
	if query, ok := args["query"].(string); ok && query != "" {
		query = strings.ToLower(query)
		filtered := make([]jira.Field, 0)
		for _, field := range fields {
			if strings.Contains(strings.ToLower(field.Name), query) ||
				strings.Contains(strings.ToLower(field.ID), query) {
				filtered = append(filtered, field)
			}
		}
		fields = filtered
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"fields": fields,
		"total":  len(fields),
	})
}

// JiraGetAllProjectsTool creates the jira_get_all_projects tool
func JiraGetAllProjectsTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_get_all_projects",
		"List all accessible Jira projects with optional expansion of project details.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"expand": mcp.NewStringProperty("Resources to expand (e.g., 'description,lead,issueTypes'). Comma-separated."),
			},
		),
		jiraGetAllProjectsHandler,
		"jira", "read",
	)
}

func jiraGetAllProjectsHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	opts := &jira.GetProjectsOptions{}

	if expand, ok := args["expand"].(string); ok && expand != "" {
		opts.Expand = strings.Split(expand, ",")
	}

	projects, err := client.GetAllProjects(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"projects": projects,
		"total":    len(projects),
	})
}

// JiraGetProjectIssuesTool creates the jira_get_project_issues tool
func JiraGetProjectIssuesTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_get_project_issues",
		"Get all issues for a specific Jira project with pagination support.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"project_key": mcp.NewStringProperty("Project key (e.g., 'PROJ')"),
				"fields": mcp.NewStringProperty("Fields to retrieve: 'essential' (default), '*all', or comma-separated field names").
					WithDefault("essential"),
				"start_at": mcp.NewIntegerProperty("Starting index for pagination (0-based)").
					WithDefault(0),
				"max_results": mcp.NewIntegerProperty("Maximum number of results to return (default 50)").
					WithDefault(50),
			},
			"project_key",
		),
		jiraGetProjectIssuesHandler,
		"jira", "read",
	)
}

func jiraGetProjectIssuesHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	projectKey, ok := args["project_key"].(string)
	if !ok || projectKey == "" {
		return nil, fmt.Errorf("project_key is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	opts := &jira.SearchOptions{
		StartAt:    getIntArg(args, "start_at", 0),
		MaxResults: getIntArg(args, "max_results", 50),
	}

	if fields, ok := args["fields"].(string); ok && fields != "" {
		if fields == "*all" {
			opts.Fields = []string{"*all"}
		} else if fields != "essential" {
			opts.Fields = strings.Split(fields, ",")
		}
	}

	result, err := client.GetProjectIssues(ctx, projectKey, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get project issues: %w", err)
	}

	return mcp.NewJSONResult(result)
}

// JiraGetProjectVersionsTool creates the jira_get_project_versions tool
func JiraGetProjectVersionsTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_get_project_versions",
		"Get all fix versions for a Jira project.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"project_key": mcp.NewStringProperty("Project key (e.g., 'PROJ')"),
			},
			"project_key",
		),
		jiraGetProjectVersionsHandler,
		"jira", "read",
	)
}

func jiraGetProjectVersionsHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	projectKey, ok := args["project_key"].(string)
	if !ok || projectKey == "" {
		return nil, fmt.Errorf("project_key is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	versions, err := client.GetProjectVersions(ctx, projectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get project versions: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"versions": versions,
		"total":    len(versions),
	})
}

// JiraGetTransitionsTool creates the jira_get_transitions tool
func JiraGetTransitionsTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_get_transitions",
		"Get available status transitions for a Jira issue.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"issue_key": mcp.NewStringProperty("Issue key (e.g., 'PROJ-123')"),
			},
			"issue_key",
		),
		jiraGetTransitionsHandler,
		"jira", "read",
	)
}

func jiraGetTransitionsHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	issueKey, ok := args["issue_key"].(string)
	if !ok || issueKey == "" {
		return nil, fmt.Errorf("issue_key is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	transitions, err := client.GetTransitions(ctx, issueKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get transitions: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"transitions": transitions,
		"total":       len(transitions),
	})
}

// JiraGetWorklogTool creates the jira_get_worklog tool
func JiraGetWorklogTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_get_worklog",
		"Get worklog entries for a Jira issue (time tracking).",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"issue_key": mcp.NewStringProperty("Issue key (e.g., 'PROJ-123')"),
			},
			"issue_key",
		),
		jiraGetWorklogHandler,
		"jira", "read",
	)
}

func jiraGetWorklogHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	issueKey, ok := args["issue_key"].(string)
	if !ok || issueKey == "" {
		return nil, fmt.Errorf("issue_key is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	worklogs, err := client.GetWorklogs(ctx, issueKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get worklogs: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"worklogs": worklogs,
		"total":    len(worklogs),
	})
}

// JiraGetAgileBoardsTool creates the jira_get_agile_boards tool
func JiraGetAgileBoardsTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_get_agile_boards",
		"Get Jira agile boards (Scrum/Kanban) with optional filtering.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"project_key": mcp.NewStringProperty("Filter boards by project key"),
				"board_type":  mcp.NewStringProperty("Filter by board type: 'scrum', 'kanban', or 'simple'"),
				"name":        mcp.NewStringProperty("Filter boards by name (partial match)"),
			},
		),
		jiraGetAgileBoardsHandler,
		"jira", "read",
	)
}

func jiraGetAgileBoardsHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	opts := &jira.GetBoardsOptions{}

	if projectKey, ok := args["project_key"].(string); ok && projectKey != "" {
		opts.ProjectKeyOrID = projectKey
	}

	if boardType, ok := args["board_type"].(string); ok && boardType != "" {
		opts.BoardType = boardType
	}

	if name, ok := args["name"].(string); ok && name != "" {
		opts.Name = name
	}

	boards, err := client.GetBoards(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get boards: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"boards": boards,
		"total":  len(boards),
	})
}

// JiraGetBoardIssuesTool creates the jira_get_board_issues tool
func JiraGetBoardIssuesTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_get_board_issues",
		"Get issues linked to a specific Jira agile board.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"board_id": mcp.NewIntegerProperty("Board ID"),
				"start_at": mcp.NewIntegerProperty("Starting index for pagination (0-based)").
					WithDefault(0),
				"max_results": mcp.NewIntegerProperty("Maximum number of results to return (default 50)").
					WithDefault(50),
			},
			"board_id",
		),
		jiraGetBoardIssuesHandler,
		"jira", "read",
	)
}

func jiraGetBoardIssuesHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	boardID := getIntArg(args, "board_id", 0)
	if boardID == 0 {
		return nil, fmt.Errorf("board_id is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	opts := &jira.SearchOptions{
		StartAt:    getIntArg(args, "start_at", 0),
		MaxResults: getIntArg(args, "max_results", 50),
	}

	result, err := client.GetBoardIssues(ctx, boardID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get board issues: %w", err)
	}

	return mcp.NewJSONResult(result)
}

// JiraGetSprintsFromBoardTool creates the jira_get_sprints_from_board tool
func JiraGetSprintsFromBoardTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_get_sprints_from_board",
		"Get sprints from a Jira agile board, optionally filtered by state.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"board_id": mcp.NewIntegerProperty("Board ID"),
				"state":    mcp.NewStringProperty("Filter by sprint state: 'future', 'active', or 'closed'"),
			},
			"board_id",
		),
		jiraGetSprintsFromBoardHandler,
		"jira", "read",
	)
}

func jiraGetSprintsFromBoardHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	boardID := getIntArg(args, "board_id", 0)
	if boardID == 0 {
		return nil, fmt.Errorf("board_id is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	state := ""
	if s, ok := args["state"].(string); ok {
		state = s
	}

	sprints, err := client.GetBoardSprints(ctx, boardID, state)
	if err != nil {
		return nil, fmt.Errorf("failed to get sprints: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"sprints": sprints,
		"total":   len(sprints),
	})
}

// JiraGetSprintIssuesTool creates the jira_get_sprint_issues tool
func JiraGetSprintIssuesTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_get_sprint_issues",
		"Get issues in a specific sprint.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"sprint_id": mcp.NewIntegerProperty("Sprint ID"),
				"start_at": mcp.NewIntegerProperty("Starting index for pagination (0-based)").
					WithDefault(0),
				"max_results": mcp.NewIntegerProperty("Maximum number of results to return (default 50)").
					WithDefault(50),
			},
			"sprint_id",
		),
		jiraGetSprintIssuesHandler,
		"jira", "read",
	)
}

func jiraGetSprintIssuesHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	sprintID := getIntArg(args, "sprint_id", 0)
	if sprintID == 0 {
		return nil, fmt.Errorf("sprint_id is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	opts := &jira.SearchOptions{
		StartAt:    getIntArg(args, "start_at", 0),
		MaxResults: getIntArg(args, "max_results", 50),
	}

	result, err := client.GetSprintIssues(ctx, sprintID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get sprint issues: %w", err)
	}

	return mcp.NewJSONResult(result)
}

// JiraGetIssueLinkTypesTool creates the jira_get_issue_link_types tool
func JiraGetIssueLinkTypesTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_get_issue_link_types",
		"Get all available issue link types (e.g., 'Blocks', 'Relates to', 'Duplicates').",
		mcp.NewInputSchema(nil),
		jiraGetIssueLinkTypesHandler,
		"jira", "read",
	)
}

func jiraGetIssueLinkTypesHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	linkTypes, err := client.GetIssueLinkTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue link types: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"linkTypes": linkTypes,
		"total":     len(linkTypes),
	})
}

// JiraGetUserProfileTool creates the jira_get_user_profile tool
func JiraGetUserProfileTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_get_user_profile",
		"Get user profile information. For Cloud, use account_id. For Server/DC, use username.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"account_id": mcp.NewStringProperty("User account ID (Cloud)"),
				"username":   mcp.NewStringProperty("Username (Server/DC)"),
			},
		),
		jiraGetUserProfileHandler,
		"jira", "read",
	)
}

func jiraGetUserProfileHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	var user *jira.User
	var err error

	if accountID, ok := args["account_id"].(string); ok && accountID != "" {
		user, err = client.GetUser(ctx, accountID)
	} else if username, ok := args["username"].(string); ok && username != "" {
		user, err = client.GetUser(ctx, username)
	} else {
		return nil, fmt.Errorf("either account_id (Cloud) or username (Server/DC) is required")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return mcp.NewJSONResult(user)
}

// Helper function to get integer argument with default
func getIntArg(args map[string]interface{}, key string, defaultVal int) int {
	if val, ok := args[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return defaultVal
}
