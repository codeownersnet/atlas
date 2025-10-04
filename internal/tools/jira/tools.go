package jira

import (
	"context"
	"fmt"

	"github.com/codeownersnet/atlas/internal/mcp"
	"github.com/codeownersnet/atlas/pkg/atlassian/jira"
)

// Context key for storing Jira client
type contextKey string

const jiraClientKey contextKey = "jira_client"

// WithJiraClient adds a Jira client to the context
func WithJiraClient(ctx context.Context, client *jira.Client) context.Context {
	return context.WithValue(ctx, jiraClientKey, client)
}

// GetJiraClient retrieves the Jira client from the context
func GetJiraClient(ctx context.Context) *jira.Client {
	client, ok := ctx.Value(jiraClientKey).(*jira.Client)
	if !ok {
		return nil
	}
	return client
}

// RegisterJiraTools registers all Jira tools with the MCP server
func RegisterJiraTools(server *mcp.Server) error {
	tools := []struct {
		name string
		tool *mcp.ToolDefinition
	}{
		// Read operations
		{"jira_get_issue", JiraGetIssueTool()},
		{"jira_search", JiraSearchTool()},
		{"jira_search_fields", JiraSearchFieldsTool()},
		{"jira_get_all_projects", JiraGetAllProjectsTool()},
		{"jira_get_project_issues", JiraGetProjectIssuesTool()},
		{"jira_get_project_versions", JiraGetProjectVersionsTool()},
		{"jira_get_transitions", JiraGetTransitionsTool()},
		{"jira_get_worklog", JiraGetWorklogTool()},
		{"jira_get_agile_boards", JiraGetAgileBoardsTool()},
		{"jira_get_board_issues", JiraGetBoardIssuesTool()},
		{"jira_get_sprints_from_board", JiraGetSprintsFromBoardTool()},
		{"jira_get_sprint_issues", JiraGetSprintIssuesTool()},
		{"jira_get_issue_link_types", JiraGetIssueLinkTypesTool()},
		{"jira_get_user_profile", JiraGetUserProfileTool()},

		// Write operations
		{"jira_create_issue", JiraCreateIssueTool()},
		{"jira_update_issue", JiraUpdateIssueTool()},
		{"jira_delete_issue", JiraDeleteIssueTool()},
		{"jira_add_comment", JiraAddCommentTool()},
		{"jira_transition_issue", JiraTransitionIssueTool()},
		{"jira_add_worklog", JiraAddWorklogTool()},
		{"jira_link_to_epic", JiraLinkToEpicTool()},
		{"jira_create_issue_link", JiraCreateIssueLinkTool()},
		{"jira_create_remote_issue_link", JiraCreateRemoteIssueLinkTool()},
		{"jira_remove_issue_link", JiraRemoveIssueLinkTool()},
		{"jira_create_sprint", JiraCreateSprintTool()},
		{"jira_update_sprint", JiraUpdateSprintTool()},
		{"jira_create_version", JiraCreateVersionTool()},
		{"jira_batch_create_issues", JiraBatchCreateIssuesTool()},
		{"jira_batch_create_versions", JiraBatchCreateVersionsTool()},
	}

	for _, t := range tools {
		if err := server.RegisterTool(t.tool); err != nil {
			return fmt.Errorf("failed to register %s: %w", t.name, err)
		}
	}

	return nil
}
