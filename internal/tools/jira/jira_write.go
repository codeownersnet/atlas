package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/codeownersnet/atlas/internal/mcp"
	"github.com/codeownersnet/atlas/pkg/atlassian/jira"
)

// JiraCreateIssueTool creates the jira_create_issue tool
func JiraCreateIssueTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_create_issue",
		"Create a new Jira issue. Requires project key, issue type, and summary at minimum. Supports custom fields and Epic linking.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"project_key": mcp.NewStringProperty("Project key (e.g., 'PROJ')"),
				"issue_type":  mcp.NewStringProperty("Issue type name (e.g., 'Bug', 'Story', 'Task')"),
				"summary":     mcp.NewStringProperty("Issue summary/title"),
				"description": mcp.NewStringProperty("Issue description. Supports rich Markdown formatting: ## headings, **bold**, *italic*, `code`, ~~strikethrough~~, ++underline++, links [text](url), lists, tables, code blocks (```lang```). Blockquotes (> text), panels ([info], [warning], [error], [success]), expand sections (<details>Title</details>), mentions (@username), status ([status:Done]), and emoji (:smile:) are also supported. Jira wiki markup (h2., *bold*, {code}, etc.) is auto-converted."),
				"fields":      mcp.NewStringProperty("Additional fields as JSON object (e.g., '{\"priority\": {\"name\": \"High\"}, \"labels\": [\"bug\"]}'). Use for custom fields and standard fields."),
			},
			"project_key", "issue_type", "summary",
		),
		jiraCreateIssueHandler,
		"jira", "write",
	)
}

func jiraCreateIssueHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	projectKey, ok := args["project_key"].(string)
	if !ok || projectKey == "" {
		return nil, fmt.Errorf("project_key is required")
	}

	issueType, ok := args["issue_type"].(string)
	if !ok || issueType == "" {
		return nil, fmt.Errorf("issue_type is required")
	}

	summary, ok := args["summary"].(string)
	if !ok || summary == "" {
		return nil, fmt.Errorf("summary is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	// Build fields map
	fields := map[string]interface{}{
		"project": map[string]string{
			"key": projectKey,
		},
		"issuetype": map[string]string{
			"name": issueType,
		},
		"summary": summary,
	}

	// Add description if provided
	if description, ok := args["description"].(string); ok && description != "" {
		fields["description"] = description
	}

	// Parse additional fields if provided
	if fieldsJSON, ok := args["fields"].(string); ok && fieldsJSON != "" {
		var additionalFields map[string]interface{}
		if err := json.Unmarshal([]byte(fieldsJSON), &additionalFields); err != nil {
			return nil, fmt.Errorf("invalid fields JSON: %w", err)
		}
		// Merge additional fields
		for k, v := range additionalFields {
			fields[k] = v
		}
	}

	issue, err := client.CreateIssue(ctx, fields)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"key":     issue.Key,
		"id":      issue.ID,
		"self":    issue.Self,
		"message": fmt.Sprintf("Successfully created issue %s", issue.Key),
	})
}

// JiraUpdateIssueTool creates the jira_update_issue tool
func JiraUpdateIssueTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_update_issue",
		"Update an existing Jira issue. Can update any field including custom fields. Description field supports rich Markdown formatting: ## headings, **bold**, *italic*, `code`, ~~strikethrough~~, ++underline++, links, lists, tables, code blocks. Blockquotes (> text), panels ([info], [warning], [error], [success]), expand sections (<details>Title</details>), mentions (@username), status ([status:Done]), and emoji (:smile:) are also supported.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"issue_key": mcp.NewStringProperty("Issue key (e.g., 'PROJ-123')"),
				"fields":    mcp.NewStringProperty("Fields to update as JSON object (e.g., '{\"summary\": \"New title\", \"priority\": {\"name\": \"High\"}}')"),
				"update":    mcp.NewStringProperty("Update operations as JSON object (e.g., '{\"labels\": [{\"add\": \"new-label\"}]}')"),
			},
			"issue_key",
		),
		jiraUpdateIssueHandler,
		"jira", "write",
	)
}

func jiraUpdateIssueHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	issueKey, ok := args["issue_key"].(string)
	if !ok || issueKey == "" {
		return nil, fmt.Errorf("issue_key is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	var fields map[string]interface{}
	var update map[string]interface{}

	// Parse fields JSON
	if fieldsJSON, ok := args["fields"].(string); ok && fieldsJSON != "" {
		if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
			return nil, fmt.Errorf("invalid fields JSON: %w", err)
		}
	}

	// Parse update JSON
	if updateJSON, ok := args["update"].(string); ok && updateJSON != "" {
		if err := json.Unmarshal([]byte(updateJSON), &update); err != nil {
			return nil, fmt.Errorf("invalid update JSON: %w", err)
		}
	}

	if fields == nil && update == nil {
		return nil, fmt.Errorf("either fields or update must be provided")
	}

	err := client.UpdateIssue(ctx, issueKey, fields, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update issue: %w", err)
	}

	return mcp.NewSuccessResult(fmt.Sprintf("Successfully updated issue %s", issueKey)), nil
}

// JiraDeleteIssueTool creates the jira_delete_issue tool
func JiraDeleteIssueTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_delete_issue",
		"Delete a Jira issue. Use with caution as this action cannot be undone.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"issue_key":       mcp.NewStringProperty("Issue key (e.g., 'PROJ-123')"),
				"delete_subtasks": mcp.NewBooleanProperty("Whether to delete subtasks (default: false)").WithDefault(false),
			},
			"issue_key",
		),
		jiraDeleteIssueHandler,
		"jira", "write",
	)
}

func jiraDeleteIssueHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	issueKey, ok := args["issue_key"].(string)
	if !ok || issueKey == "" {
		return nil, fmt.Errorf("issue_key is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	deleteSubtasks := false
	if val, ok := args["delete_subtasks"].(bool); ok {
		deleteSubtasks = val
	}

	err := client.DeleteIssue(ctx, issueKey, deleteSubtasks)
	if err != nil {
		return nil, fmt.Errorf("failed to delete issue: %w", err)
	}

	return mcp.NewSuccessResult(fmt.Sprintf("Successfully deleted issue %s", issueKey)), nil
}

// JiraAddCommentTool creates the jira_add_comment tool
func JiraAddCommentTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_add_comment",
		"Add a comment to a Jira issue. Supports rich Markdown formatting: ## headings, **bold**, *italic*, `code`, ~~strikethrough~~, ++underline++, links [text](url), lists, tables, code blocks. Blockquotes (> text), panels ([info], [warning], [error], [success]), expand sections (<details>Title</details>), mentions (@username), status ([status:Done]), and emoji (:smile:) are also supported. Jira wiki markup is auto-converted.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"issue_key": mcp.NewStringProperty("Issue key (e.g., 'PROJ-123')"),
				"body":      mcp.NewStringProperty("Comment text/body"),
			},
			"issue_key", "body",
		),
		jiraAddCommentHandler,
		"jira", "write",
	)
}

func jiraAddCommentHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	issueKey, ok := args["issue_key"].(string)
	if !ok || issueKey == "" {
		return nil, fmt.Errorf("issue_key is required")
	}

	body, ok := args["body"].(string)
	if !ok || body == "" {
		return nil, fmt.Errorf("body is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	comment, err := client.AddComment(ctx, issueKey, body, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"id":      comment.ID,
		"message": fmt.Sprintf("Successfully added comment to issue %s", issueKey),
	})
}

// JiraTransitionIssueTool creates the jira_transition_issue tool
func JiraTransitionIssueTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_transition_issue",
		"Transition a Jira issue to a different status (e.g., 'In Progress', 'Done'). Use jira_get_transitions to see available transitions.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"issue_key":     mcp.NewStringProperty("Issue key (e.g., 'PROJ-123')"),
				"transition_id": mcp.NewStringProperty("Transition ID or name"),
				"comment":       mcp.NewStringProperty("Optional comment to add with the transition"),
			},
			"issue_key", "transition_id",
		),
		jiraTransitionIssueHandler,
		"jira", "write",
	)
}

func jiraTransitionIssueHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	issueKey, ok := args["issue_key"].(string)
	if !ok || issueKey == "" {
		return nil, fmt.Errorf("issue_key is required")
	}

	transitionID, ok := args["transition_id"].(string)
	if !ok || transitionID == "" {
		return nil, fmt.Errorf("transition_id is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	// Build fields map for optional comment
	fields := make(map[string]interface{})
	if c, ok := args["comment"].(string); ok && c != "" {
		fields["comment"] = []map[string]string{
			{"add": c},
		}
	}

	err := client.TransitionIssue(ctx, issueKey, transitionID, fields)
	if err != nil {
		return nil, fmt.Errorf("failed to transition issue: %w", err)
	}

	return mcp.NewSuccessResult(fmt.Sprintf("Successfully transitioned issue %s", issueKey)), nil
}

// JiraAddWorklogTool creates the jira_add_worklog tool
func JiraAddWorklogTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_add_worklog",
		"Add a worklog entry to a Jira issue for time tracking. Time spent should be in Jira format (e.g., '2h 30m', '1d', '3w').",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"issue_key":  mcp.NewStringProperty("Issue key (e.g., 'PROJ-123')"),
				"time_spent": mcp.NewStringProperty("Time spent in Jira format (e.g., '2h 30m', '1d', '3w')"),
				"comment":    mcp.NewStringProperty("Work description/comment"),
				"started":    mcp.NewStringProperty("When the work was started (ISO 8601 format, e.g., '2025-01-15T10:00:00.000+0000'). Defaults to now."),
			},
			"issue_key", "time_spent",
		),
		jiraAddWorklogHandler,
		"jira", "write",
	)
}

func jiraAddWorklogHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	issueKey, ok := args["issue_key"].(string)
	if !ok || issueKey == "" {
		return nil, fmt.Errorf("issue_key is required")
	}

	timeSpent, ok := args["time_spent"].(string)
	if !ok || timeSpent == "" {
		return nil, fmt.Errorf("time_spent is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	// Convert time spent to seconds first
	timeSpentSeconds, err := parseJiraTime(timeSpent)
	if err != nil {
		return nil, fmt.Errorf("invalid time_spent format: %w", err)
	}

	// Build worklog request
	req := &jira.CreateWorklogRequest{
		TimeSpentSeconds: timeSpentSeconds,
	}

	if c, ok := args["comment"].(string); ok && c != "" {
		req.Comment = c
	}

	if s, ok := args["started"].(string); ok && s != "" {
		req.Started = s
	} else {
		// Default to current time in ISO 8601 format
		req.Started = time.Now().Format("2006-01-02T15:04:05.000-0700")
	}

	worklog, err := client.AddWorklog(ctx, issueKey, req)
	if err != nil {
		return nil, fmt.Errorf("failed to add worklog: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"id":      worklog.ID,
		"message": fmt.Sprintf("Successfully added worklog to issue %s", issueKey),
	})
}

// JiraLinkToEpicTool creates the jira_link_to_epic tool
func JiraLinkToEpicTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_link_to_epic",
		"Link a Jira issue to an Epic. Note: Epic linking works differently in Cloud vs Server/DC.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"issue_key": mcp.NewStringProperty("Issue key to link (e.g., 'PROJ-123')"),
				"epic_key":  mcp.NewStringProperty("Epic key to link to (e.g., 'PROJ-100')"),
			},
			"issue_key", "epic_key",
		),
		jiraLinkToEpicHandler,
		"jira", "write",
	)
}

func jiraLinkToEpicHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	issueKey, ok := args["issue_key"].(string)
	if !ok || issueKey == "" {
		return nil, fmt.Errorf("issue_key is required")
	}

	epicKey, ok := args["epic_key"].(string)
	if !ok || epicKey == "" {
		return nil, fmt.Errorf("epic_key is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	err := client.LinkToEpic(ctx, issueKey, epicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to link to epic: %w", err)
	}

	return mcp.NewSuccessResult(fmt.Sprintf("Successfully linked issue %s to epic %s", issueKey, epicKey)), nil
}

// JiraCreateIssueLinkTool creates the jira_create_issue_link tool
func JiraCreateIssueLinkTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_create_issue_link",
		"Create a link between two Jira issues (e.g., 'Blocks', 'Relates to', 'Duplicates'). Use jira_get_issue_link_types to see available link types.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"from_key":  mcp.NewStringProperty("Source issue key (e.g., 'PROJ-123')"),
				"to_key":    mcp.NewStringProperty("Target issue key (e.g., 'PROJ-456')"),
				"link_type": mcp.NewStringProperty("Link type name (e.g., 'Blocks', 'Relates to', 'Duplicates')"),
				"comment":   mcp.NewStringProperty("Optional comment for the link"),
			},
			"from_key", "to_key", "link_type",
		),
		jiraCreateIssueLinkHandler,
		"jira", "write",
	)
}

func jiraCreateIssueLinkHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	fromKey, ok := args["from_key"].(string)
	if !ok || fromKey == "" {
		return nil, fmt.Errorf("from_key is required")
	}

	toKey, ok := args["to_key"].(string)
	if !ok || toKey == "" {
		return nil, fmt.Errorf("to_key is required")
	}

	linkType, ok := args["link_type"].(string)
	if !ok || linkType == "" {
		return nil, fmt.Errorf("link_type is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	// Build optional comment
	var commentObj *jira.Comment
	if c, ok := args["comment"].(string); ok && c != "" {
		commentObj = &jira.Comment{
			Body: jira.NewDescription(c),
		}
	}

	// Use the helper method that looks up the link type by name
	_, err := client.CreateIssueLinkByName(ctx, linkType, fromKey, toKey, commentObj)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue link: %w", err)
	}

	return mcp.NewSuccessResult(fmt.Sprintf("Successfully linked %s to %s with type '%s'", fromKey, toKey, linkType)), nil
}

// JiraCreateRemoteIssueLinkTool creates the jira_create_remote_issue_link tool
func JiraCreateRemoteIssueLinkTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_create_remote_issue_link",
		"Create a remote/external link from a Jira issue to an external URL or resource.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"issue_key": mcp.NewStringProperty("Issue key (e.g., 'PROJ-123')"),
				"url":       mcp.NewStringProperty("External URL to link to"),
				"title":     mcp.NewStringProperty("Link title/description"),
			},
			"issue_key", "url", "title",
		),
		jiraCreateRemoteIssueLinkHandler,
		"jira", "write",
	)
}

func jiraCreateRemoteIssueLinkHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	issueKey, ok := args["issue_key"].(string)
	if !ok || issueKey == "" {
		return nil, fmt.Errorf("issue_key is required")
	}

	url, ok := args["url"].(string)
	if !ok || url == "" {
		return nil, fmt.Errorf("url is required")
	}

	title, ok := args["title"].(string)
	if !ok || title == "" {
		return nil, fmt.Errorf("title is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	remoteLink := &jira.RemoteLink{
		Object: &jira.LinkObject{
			URL:   url,
			Title: title,
		},
	}

	result, err := client.CreateRemoteLink(ctx, issueKey, remoteLink)
	if err != nil {
		return nil, fmt.Errorf("failed to create remote link: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"id":      result.ID,
		"message": fmt.Sprintf("Successfully created remote link on issue %s", issueKey),
	})
}

// JiraRemoveIssueLinkTool creates the jira_remove_issue_link tool
func JiraRemoveIssueLinkTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_remove_issue_link",
		"Remove a link between Jira issues.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"link_id": mcp.NewStringProperty("Link ID to remove"),
			},
			"link_id",
		),
		jiraRemoveIssueLinkHandler,
		"jira", "write",
	)
}

func jiraRemoveIssueLinkHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	linkID, ok := args["link_id"].(string)
	if !ok || linkID == "" {
		return nil, fmt.Errorf("link_id is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	err := client.DeleteIssueLink(ctx, linkID)
	if err != nil {
		return nil, fmt.Errorf("failed to remove issue link: %w", err)
	}

	return mcp.NewSuccessResult(fmt.Sprintf("Successfully removed link %s", linkID)), nil
}

// JiraCreateSprintTool creates the jira_create_sprint tool
func JiraCreateSprintTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_create_sprint",
		"Create a new sprint in a Jira Scrum board.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"board_id":   mcp.NewIntegerProperty("Board ID where the sprint will be created"),
				"name":       mcp.NewStringProperty("Sprint name"),
				"start_date": mcp.NewStringProperty("Sprint start date (ISO 8601 format, e.g., '2025-01-15T10:00:00.000Z')"),
				"end_date":   mcp.NewStringProperty("Sprint end date (ISO 8601 format, e.g., '2025-01-29T10:00:00.000Z')"),
				"goal":       mcp.NewStringProperty("Sprint goal/objective"),
			},
			"board_id", "name",
		),
		jiraCreateSprintHandler,
		"jira", "write",
	)
}

func jiraCreateSprintHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	boardID := getIntArg(args, "board_id", 0)
	if boardID == 0 {
		return nil, fmt.Errorf("board_id is required")
	}

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("name is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	req := &jira.CreateSprintRequest{
		Name:          name,
		OriginBoardID: boardID,
	}

	if startDate, ok := args["start_date"].(string); ok && startDate != "" {
		req.StartDate = startDate
	}

	if endDate, ok := args["end_date"].(string); ok && endDate != "" {
		req.EndDate = endDate
	}

	if goal, ok := args["goal"].(string); ok && goal != "" {
		req.Goal = goal
	}

	sprint, err := client.CreateSprint(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create sprint: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"id":      sprint.ID,
		"name":    sprint.Name,
		"state":   sprint.State,
		"message": fmt.Sprintf("Successfully created sprint '%s'", sprint.Name),
	})
}

// JiraUpdateSprintTool creates the jira_update_sprint tool
func JiraUpdateSprintTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_update_sprint",
		"Update an existing sprint. Can update name, dates, goal, and state (start/close sprint).",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"sprint_id":  mcp.NewIntegerProperty("Sprint ID"),
				"name":       mcp.NewStringProperty("New sprint name"),
				"start_date": mcp.NewStringProperty("New start date (ISO 8601 format)"),
				"end_date":   mcp.NewStringProperty("New end date (ISO 8601 format)"),
				"goal":       mcp.NewStringProperty("New sprint goal"),
				"state":      mcp.NewStringProperty("Sprint state: 'future', 'active', or 'closed'"),
			},
			"sprint_id",
		),
		jiraUpdateSprintHandler,
		"jira", "write",
	)
}

func jiraUpdateSprintHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	sprintID := getIntArg(args, "sprint_id", 0)
	if sprintID == 0 {
		return nil, fmt.Errorf("sprint_id is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	req := &jira.UpdateSprintRequest{}
	hasUpdate := false

	if name, ok := args["name"].(string); ok && name != "" {
		req.Name = name
		hasUpdate = true
	}

	if startDate, ok := args["start_date"].(string); ok && startDate != "" {
		req.StartDate = startDate
		hasUpdate = true
	}

	if endDate, ok := args["end_date"].(string); ok && endDate != "" {
		req.EndDate = endDate
		hasUpdate = true
	}

	if goal, ok := args["goal"].(string); ok && goal != "" {
		req.Goal = goal
		hasUpdate = true
	}

	if state, ok := args["state"].(string); ok && state != "" {
		req.State = state
		hasUpdate = true
	}

	if !hasUpdate {
		return nil, fmt.Errorf("at least one field to update must be provided")
	}

	sprint, err := client.UpdateSprint(ctx, sprintID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update sprint: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"id":      sprint.ID,
		"name":    sprint.Name,
		"state":   sprint.State,
		"message": fmt.Sprintf("Successfully updated sprint %d", sprintID),
	})
}

// JiraCreateVersionTool creates the jira_create_version tool
func JiraCreateVersionTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_create_version",
		"Create a new fix version in a Jira project.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"project_key":  mcp.NewStringProperty("Project key (e.g., 'PROJ')"),
				"name":         mcp.NewStringProperty("Version name (e.g., '1.0.0')"),
				"description":  mcp.NewStringProperty("Version description"),
				"release_date": mcp.NewStringProperty("Planned release date (YYYY-MM-DD format)"),
				"released":     mcp.NewBooleanProperty("Whether the version is released").WithDefault(false),
			},
			"project_key", "name",
		),
		jiraCreateVersionHandler,
		"jira", "write",
	)
}

func jiraCreateVersionHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	projectKey, ok := args["project_key"].(string)
	if !ok || projectKey == "" {
		return nil, fmt.Errorf("project_key is required")
	}

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("name is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	req := &jira.CreateVersionRequest{
		Name:    name,
		Project: projectKey,
	}

	if description, ok := args["description"].(string); ok && description != "" {
		req.Description = description
	}

	if releaseDate, ok := args["release_date"].(string); ok && releaseDate != "" {
		req.ReleaseDate = releaseDate
	}

	if released, ok := args["released"].(bool); ok {
		req.Released = released
	}

	version, err := client.CreateVersion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"id":      version.ID,
		"name":    version.Name,
		"message": fmt.Sprintf("Successfully created version '%s' in project %s", version.Name, projectKey),
	})
}

// JiraBatchCreateIssuesTool creates the jira_batch_create_issues tool
func JiraBatchCreateIssuesTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_batch_create_issues",
		"Create multiple Jira issues in a single batch operation. More efficient than creating issues one by one.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"issues": mcp.NewStringProperty("JSON array of issue definitions. Each issue should have fields object with project, issuetype, summary, etc. Example: '[{\"fields\": {\"project\": {\"key\": \"PROJ\"}, \"issuetype\": {\"name\": \"Task\"}, \"summary\": \"Issue 1\"}}, {...}]'"),
			},
			"issues",
		),
		jiraBatchCreateIssuesHandler,
		"jira", "write",
	)
}

func jiraBatchCreateIssuesHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	issuesJSON, ok := args["issues"].(string)
	if !ok || issuesJSON == "" {
		return nil, fmt.Errorf("issues is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	// Parse issues array
	var issuesArray []map[string]interface{}
	if err := json.Unmarshal([]byte(issuesJSON), &issuesArray); err != nil {
		return nil, fmt.Errorf("invalid issues JSON: %w", err)
	}

	// Extract fields from each issue
	issuesFields := make([]map[string]interface{}, len(issuesArray))
	for i, issue := range issuesArray {
		if fields, ok := issue["fields"].(map[string]interface{}); ok {
			issuesFields[i] = fields
		} else {
			return nil, fmt.Errorf("issue at index %d is missing 'fields' object", i)
		}
	}

	result, err := client.BatchCreateIssues(ctx, issuesFields)
	if err != nil {
		return nil, fmt.Errorf("failed to batch create issues: %w", err)
	}

	// Build response
	created := make([]map[string]string, 0)
	for _, issue := range result.Issues {
		created = append(created, map[string]string{
			"key": issue.Key,
			"id":  issue.ID,
		})
	}

	errors := make([]map[string]interface{}, 0)
	for _, err := range result.Errors {
		errors = append(errors, map[string]interface{}{
			"status":        err.Status,
			"failedElement": err.FailedElement,
		})
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"created": created,
		"errors":  errors,
		"message": fmt.Sprintf("Successfully created %d issues", len(created)),
	})
}

// JiraBatchCreateVersionsTool creates the jira_batch_create_versions tool
func JiraBatchCreateVersionsTool() *mcp.ToolDefinition {
	return mcp.NewTool(
		"jira_batch_create_versions",
		"Create multiple fix versions in a single batch operation.",
		mcp.NewInputSchema(
			map[string]mcp.Property{
				"project_key": mcp.NewStringProperty("Project key (e.g., 'PROJ')"),
				"versions":    mcp.NewStringProperty("JSON array of version definitions. Example: '[{\"name\": \"1.0.0\", \"description\": \"First release\"}, {\"name\": \"1.1.0\"}]'"),
			},
			"project_key", "versions",
		),
		jiraBatchCreateVersionsHandler,
		"jira", "write",
	)
}

func jiraBatchCreateVersionsHandler(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	projectKey, ok := args["project_key"].(string)
	if !ok || projectKey == "" {
		return nil, fmt.Errorf("project_key is required")
	}

	versionsJSON, ok := args["versions"].(string)
	if !ok || versionsJSON == "" {
		return nil, fmt.Errorf("versions is required")
	}

	client := GetJiraClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("Jira client not available")
	}

	// Parse versions array
	var versionsArray []map[string]interface{}
	if err := json.Unmarshal([]byte(versionsJSON), &versionsArray); err != nil {
		return nil, fmt.Errorf("invalid versions JSON: %w", err)
	}

	// Create versions one by one (Jira doesn't have a batch version endpoint)
	created := make([]map[string]interface{}, 0)
	errors := make([]string, 0)

	for i, versionData := range versionsArray {
		name, ok := versionData["name"].(string)
		if !ok || name == "" {
			errors = append(errors, fmt.Sprintf("version at index %d is missing 'name'", i))
			continue
		}

		req := &jira.CreateVersionRequest{
			Name:    name,
			Project: projectKey,
		}

		if description, ok := versionData["description"].(string); ok {
			req.Description = description
		}

		if releaseDate, ok := versionData["release_date"].(string); ok {
			req.ReleaseDate = releaseDate
		}

		if released, ok := versionData["released"].(bool); ok {
			req.Released = released
		}

		version, err := client.CreateVersion(ctx, req)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to create version '%s': %v", name, err))
			continue
		}

		created = append(created, map[string]interface{}{
			"id":   version.ID,
			"name": version.Name,
		})
	}

	return mcp.NewJSONResult(map[string]interface{}{
		"created": created,
		"errors":  errors,
		"message": fmt.Sprintf("Successfully created %d versions", len(created)),
	})
}

// parseJiraTime converts Jira time format (e.g., "2h 30m", "1d", "3w") to seconds
func parseJiraTime(timeStr string) (int, error) {
	// Regex to match time units: w (weeks), d (days), h (hours), m (minutes)
	re := regexp.MustCompile(`(\d+)([wdhm])`)
	matches := re.FindAllStringSubmatch(strings.ToLower(strings.TrimSpace(timeStr)), -1)

	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid time format: %s (expected format like '2h 30m', '1d', '3w')", timeStr)
	}

	totalSeconds := 0
	for _, match := range matches {
		value, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, fmt.Errorf("invalid number in time format: %s", match[1])
		}

		unit := match[2]
		switch unit {
		case "w":
			totalSeconds += value * 5 * 8 * 60 * 60 // 5 days * 8 hours
		case "d":
			totalSeconds += value * 8 * 60 * 60 // 8 hours
		case "h":
			totalSeconds += value * 60 * 60
		case "m":
			totalSeconds += value * 60
		default:
			return 0, fmt.Errorf("unknown time unit: %s", unit)
		}
	}

	return totalSeconds, nil
}
